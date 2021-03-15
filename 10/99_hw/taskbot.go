package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var WebhookURL = "https://vctrl-taskbot.herokuapp.com/"
var BotToken = os.Getenv("BOT_TOKEN")
var Port = ":" + os.Getenv("PORT")

type cmdHandler func(upd *tgbotapi.Update, params ...string) (map[int]string, error)

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatalf("error creating new BotAPI instance: %v", err)
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		log.Fatalf("error setting webhook: %v", err)
	}

	updates := bot.ListenForWebhook("/")
	server := &http.Server{
		Addr: Port,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to listen and serve: %+v", err)
		}
	}()

	fmt.Printf("server started at %s\n", Port)

	tb := NewTaskBot()
	for update := range updates {
		result, err := tb.ExecuteCmd(&update)
		chatID := update.Message.Chat.ID
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "error happened"))
		}

		for id, text := range result {
			bot.Send(tgbotapi.NewMessage(int64(id), text))
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	err = server.Shutdown(ctx)
	if err != nil {
		server.Close()
		return err
	}

	return nil
}

type TaskBot struct {
	Repo *TasksRepoInMemory
	Cmds map[string]cmdHandler
}

func NewTaskBot() *TaskBot {
	tb := &TaskBot{
		Repo: NewTasksRepoInMemory(),
	}

	tb.Cmds = map[string]cmdHandler{
		"/tasks":    tb.GetTasks,
		"/new":      tb.CreateTask,
		"/assign":   tb.AssignTask,
		"/unassign": tb.UnassignTask,
		"/resolve":  tb.ResolveTask,
		"/my":       tb.GetAssignedToUser,
		"/owner":    tb.GetCreatedByUser,
	}

	return tb
}

func (tb *TaskBot) ExecuteCmd(upd *tgbotapi.Update) (map[int]string, error) {
	var cmd, arg string
	text := upd.Message.Text
	// неуклюжий роутер
	if strings.HasPrefix(text, "/new ") {
		cmd = "/new"
		arg = strings.TrimPrefix(text, "/new ")
	} else {
		parts := strings.Split(text, "_")
		cmd = parts[0]
		if len(parts) > 1 {
			arg = parts[1]
		}
	}

	cmdHandler, ok := tb.Cmds[cmd]
	if !ok {
		return nil, fmt.Errorf("command is not supported")
	}

	return cmdHandler(upd, arg)
}

func (tb *TaskBot) GetTasks(upd *tgbotapi.Update, args ...string) (map[int]string, error) {
	userID := upd.Message.From.ID

	var result = make(map[int]string)
	tasks, err := tb.Repo.GetAll()

	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		result[userID] = "Нет задач"
		return result, nil
	}

	var sb strings.Builder
	for i, t := range tasks {
		if i != 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(t.String(userID, true))
	}

	result[userID] = sb.String()
	return result, nil
}

func (tb *TaskBot) CreateTask(upd *tgbotapi.Update, args ...string) (map[int]string, error) {
	info := args[0]
	task := &Task{info: info, createdBy: upd.Message.From}
	id, err := tb.Repo.Add(task)
	if err != nil {
		return nil, err
	}
	userID := upd.Message.From.ID
	result := map[int]string{userID: fmt.Sprintf("Задача \"%s\" создана, id=%d", info, id)}
	return result, nil
}

func (tb *TaskBot) AssignTask(upd *tgbotapi.Update, args ...string) (map[int]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("invalid args, no id")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing task id: %s", err.Error())
	}

	task, err := tb.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	userID := upd.Message.From.ID
	if task == nil {
		return map[int]string{userID: "Задача с таким id не найдена"}, nil
	}
	oldUsr := task.assignedTo
	ok, err := tb.Repo.AssignTo(id, upd.Message.From)
	if err != nil {
		return nil, err
	}
	if !ok {
		return map[int]string{userID: "Задача с таким id не найдена"}, nil
	}

	result := map[int]string{
		userID: fmt.Sprintf("Задача \"%s\" назначена на вас", task.info),
	}

	if oldUsr != nil && oldUsr.ID != userID {
		result[oldUsr.ID] =
			fmt.Sprintf("Задача \"%s\" назначена на @%s", task.info, upd.Message.From)
	} else if task.createdBy.ID != userID {
		result[task.createdBy.ID] =
			fmt.Sprintf("Задача \"%s\" назначена на @%s", task.info, upd.Message.From)
	}

	return result, nil
}

func (tb *TaskBot) UnassignTask(upd *tgbotapi.Update, args ...string) (map[int]string, error) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing task id: %s", err.Error())
	}

	task, err := tb.Repo.GetByID(id)
	user := upd.Message.From
	if task.assignedTo.ID != user.ID {
		return map[int]string{user.ID: "Задача не на вас"}, nil
	}

	ok, err := tb.Repo.Unassign(id, upd.Message.From)
	if err != nil {
		return nil, err
	}
	if !ok {
		return map[int]string{user.ID: "Задача с таким id не найдена"}, nil
	}
	return map[int]string{
		user.ID:           "Принято",
		task.createdBy.ID: fmt.Sprintf("Задача \"%s\" осталась без исполнителя", task.info),
	}, nil
}

func (tb *TaskBot) ResolveTask(upd *tgbotapi.Update, args ...string) (map[int]string, error) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing task id: %s", err.Error())
	}

	t, err := tb.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	userID := upd.Message.From.ID
	if t == nil {
		return map[int]string{userID: "Задача с таким id не найдена"}, nil
	}
	ok, err := tb.Repo.Delete(id)
	if err != nil {
		return nil, err
	}

	if !ok {
		return map[int]string{userID: "Упс! Задача уже удалена кем-то другим"}, nil
	}

	return map[int]string{
		userID:         fmt.Sprintf("Задача \"%s\" выполнена", t.info),
		t.createdBy.ID: fmt.Sprintf("Задача \"%s\" выполнена @%s", t.info, upd.Message.From),
	}, nil
}

func (tb *TaskBot) GetAssignedToUser(upd *tgbotapi.Update, args ...string) (map[int]string, error) {
	tasks, err := tb.Repo.GetByAssigneeID(upd.Message.From.ID)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	for i, t := range tasks {
		if i != 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(t.String(upd.Message.From.ID, false))
	}

	return map[int]string{upd.Message.From.ID: sb.String()}, nil
}

func (tb *TaskBot) GetCreatedByUser(upd *tgbotapi.Update, args ...string) (map[int]string, error) {
	tasks, err := tb.Repo.GetByAuthorID(upd.Message.From.ID)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	for i, t := range tasks {
		if i != 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(t.String(upd.Message.From.ID, false))
	}

	return map[int]string{upd.Message.From.ID: sb.String()}, nil
}
