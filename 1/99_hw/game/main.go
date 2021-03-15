package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Command func(*World, ...string) string

type Location struct {
	items   [][]string
	places  []string
	clothes map[string]string

	locations []string

	locks map[string]bool

	showTasks bool

	hidden         map[string]string
	description    string
	welcome        string
	triggers       map[string]func(*World) string
	noItemsMessage string
}

func (l *Location) getAvailableLocations() string {
	return fmt.Sprintf("можно пройти - %s", strings.Join(l.locations, ", "))
}

func (l *Location) getItems() string {
	res := make([]string, 0)

	// тут не нужно отображать пустые "" записи(удаление реализовано именно так)
	for i := 0; i < len(l.items); i++ {
		notEmpty := make([]string, 0)
		for _, item := range l.items[i] {
			if item != "" {
				notEmpty = append(notEmpty, item)
			}
		}

		if len(notEmpty) != 0 {
			res = append(res, strings.Join([]string{l.places[i], strings.Join(notEmpty, ", ")}, ": "))
		}
	}

	for item, place := range l.clothes {
		res = append(res, fmt.Sprintf("%s: %s", place, item))
	}

	if len(res) == 0 {
		return l.noItemsMessage
	}

	return strings.Join(res, ", ")
}

func (l *Location) openLocation(name string) {
	l.locks[name] = true
}

type Player struct {
	items map[string]bool
}

type World struct {
	currentLocation *Location
	locations       map[string]*Location
	items           map[string]string
	places          []string
	accomplished    map[string]bool
	tasks           []string
	inventoryOn     bool
}

func (w *World) getCurrentTasks() string {
	notEmptyTasks := make([]string, 0)
	for _, t := range w.tasks {
		if t != "" {
			notEmptyTasks = append(notEmptyTasks, t)
		}
	}

	return fmt.Sprintf("надо %s", strings.Join(notEmptyTasks, " и "))
}

func (w *World) accomplish(task string) {
	delete(w.accomplished, task)
	for i, v := range w.tasks {
		if v == task {
			w.tasks[i] = ""
		}
	}
}

func main() {
	initGame()

	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		fmt.Println(handleCommand(s.Text()))
	}
}

var world *World
var commands map[string]Command

func initGame() {
	commands = initCommands()

	room := &Location{
		items: [][]string{
			{"ключи", "конспекты"},
		},
		places: []string{
			"на столе",
		},
		clothes: map[string]string{
			"рюкзак": "на стуле",
		},
		locks: map[string]bool{
			"коридор": true,
		},
		locations: []string{"коридор"},
		triggers: map[string]func(*World) string{
			"взять конспекты": func(w *World) string {
				w.accomplish("собрать рюкзак")
				return ""
			},
			"надеть рюкзак": func(w *World) string {
				w.inventoryOn = true
				return ""
			},
		},
		welcome:        "ты в своей комнате",
		noItemsMessage: "пустая комната",
	}

	kitchen := &Location{
		items: [][]string{
			{"чай"},
		},
		places: []string{"на столе"},
		locks: map[string]bool{
			"коридор": true,
		},
		locations:   []string{"коридор"},
		description: "ты находишься на кухне",
		welcome:     "кухня, ничего интересного",
		showTasks:   true,
	}

	corridor := &Location{
		locks: map[string]bool{
			"кухня":   true,
			"комната": true,
			"улица":   false,
		},
		locations: []string{"кухня", "комната", "улица"},
		hidden: map[string]string{
			"дверь": "дверь на улицу",
		},
		triggers: map[string]func(*World) string{
			"применить ключи дверь": func(w *World) string {
				w.currentLocation.openLocation("улица")
				return "дверь открыта"
			},
		},
		welcome: "ничего интересного",
	}

	street := &Location{
		locks: map[string]bool{
			"домой": true,
		},
		locations: []string{"домой"},
		welcome:   "на улице весна",
	}

	world = &World{currentLocation: kitchen,
		locations: map[string]*Location{"комната": room, "кухня": kitchen, "коридор": corridor, "улица": street},
		items:     make(map[string]string),
		accomplished: map[string]bool{"собрать рюкзак": false,
			"идти в универ": false},
		tasks: []string{"собрать рюкзак", "идти в универ"},
	}
}

func handleCommand(command string) string {
	parts := strings.Split(command, " ")
	executeCmd, ok := commands[parts[0]]
	if !ok {
		return "неизвестная команда"
	}

	res := executeCmd(world, parts[1:]...)

	return res
}

func initCommands() map[string]Command {
	walk := func(w *World, args ...string) string {
		target := args[0]

		isOpen, ok := w.currentLocation.locks[target]
		if !ok {
			return fmt.Sprintf("нет пути в %s", target)
		}

		if isOpen {
			next := w.locations[target]
			w.currentLocation = next
			return fmt.Sprintf("%s. %s", next.welcome, next.getAvailableLocations())
		}

		return "дверь закрыта"
	}

	wear := func(w *World, args ...string) string {
		if len(args) < 1 {
			return "недостаточно аргументов для выполнения команды"
		}

		item := args[0]
		_, ok := w.currentLocation.clothes[item]
		if !ok {
			return "нет такого"
		}

		w.items[item] = item
		delete(w.currentLocation.clothes, item)

		trigger, ok := w.currentLocation.triggers["надеть "+item]
		if ok {
			trigger(w)
		}

		return fmt.Sprintf("вы надели: %s", item)
	}

	take := func(w *World, args ...string) string {
		if len(args) < 1 {
			return "недостаточно аргументов для выполнения команды"
		}

		if !w.inventoryOn {
			return "некуда класть"
		}

		item := args[0]

		// можно создавать один раз и хранить в структуре
		itemsMap := make(map[string]bool)
		for _, placeItems := range w.currentLocation.items {
			for _, j := range placeItems {
				itemsMap[j] = true
			}
		}

		_, ok := itemsMap[item]
		if !ok {
			return "нет такого"
		}

		w.items[item] = item
		items := w.currentLocation.items

		// удаляем путём присваивания пустой строки, которая потом будет не отображена в результате
		for i := 0; i < len(items); i++ {
			for j := 0; j < len(items[i]); j++ {
				if items[i][j] == item {
					items[i][j] = ""
				}
			}
		}

		trigger, ok := w.currentLocation.triggers["взять "+item]
		if ok {
			trigger(w)
		}

		return fmt.Sprintf("предмет добавлен в инвентарь: %s", item)
	}

	use := func(w *World, args ...string) string {
		if len(args) < 2 {
			return "недостаточно аргументов для выполнения команды"
		}

		item1 := args[0]
		_, ok := w.items[item1]
		if !ok {
			return fmt.Sprintf("нет предмета в инвентаре - %s", item1)
		}

		item2 := args[1]

		itemsMap := make(map[string]bool)
		for _, placeItems := range w.currentLocation.items {
			for _, j := range placeItems {
				itemsMap[j] = true
			}
		}

		_, ok = itemsMap[item2]
		if !ok {
			_, ok = w.currentLocation.hidden[item2]
			if !ok {
				return "не к чему применить"
			}
		}

		l := w.currentLocation
		trigger, ok := l.triggers["применить "+item1+" "+item2]
		if !ok {
			return "не к чему применить"
		}

		return trigger(w)
	}

	lookAround := func(w *World, args ...string) string {
		description := w.currentLocation.description
		items := w.currentLocation.getItems()
		var tasks string
		if w.currentLocation.showTasks {
			tasks = w.getCurrentTasks()
		} else {
			tasks = ""
		}

		notEmpty := make([]string, 0, 3)
		for _, info := range []string{description, items, tasks} {
			if info != "" {
				notEmpty = append(notEmpty, info)
			}
		}

		return fmt.Sprintf("%s. %s", strings.Join(notEmpty, ", "), w.currentLocation.getAvailableLocations())
	}

	return map[string]Command{
		"идти":        walk,
		"надеть":      wear,
		"взять":       take,
		"применить":   use,
		"осмотреться": lookAround,
	}
}
