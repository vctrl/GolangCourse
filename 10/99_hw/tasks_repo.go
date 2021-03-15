package main

import (
	"fmt"
	"sync"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type TasksRepoInMemory struct {
	lastID int
	tasks  []*Task
	mu     *sync.Mutex
}

func NewTasksRepoInMemory() *TasksRepoInMemory {
	return &TasksRepoInMemory{tasks: make([]*Task, 0, 10), mu: &sync.Mutex{}}
}

type Task struct {
	ID         int
	info       string
	assignedTo *tgbotapi.User
	createdBy  *tgbotapi.User
}

func (t *Task) String(currUserID int, includeAssignee bool) string {
	res := fmt.Sprintf("%d. %s by @%s", t.ID, t.info, t.createdBy.String())

	if includeAssignee && t.assignedTo != nil {
		var assignee string
		if t.assignedTo.ID == currUserID {
			assignee = "я"
		} else {
			assignee = fmt.Sprintf("@%s", t.assignedTo.UserName)
		}
		res += fmt.Sprintf("\nassignee: %s", assignee)
	}

	res += t.getAvailableCmds(currUserID)
	return res
}

func (t *Task) getAvailableCmds(currUserID int) string {
	if t.assignedTo == nil {
		return fmt.Sprintf("\n/assign_%d", t.ID)
	}
	if t.assignedTo != nil && t.assignedTo.ID == currUserID {
		return fmt.Sprintf("\n/unassign_%d /resolve_%d", t.ID, t.ID)
	}

	return ""
}

func (r *TasksRepoInMemory) GetAll() ([]*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tasks, nil
}

func (r *TasksRepoInMemory) GetByUserID(id int) ([]*Task, error) {
	r.mu.Lock()
	res := make([]*Task, 0)
	for _, t := range r.tasks {
		if t.createdBy.ID == id {
			res = append(res, t)
		}
	}

	r.mu.Unlock()
	return res, nil
}

func (r *TasksRepoInMemory) Add(task *Task) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task.ID = r.lastID + 1
	r.tasks = append(r.tasks, task)
	r.lastID++

	return task.ID, nil
}

func (r *TasksRepoInMemory) AssignTo(taskID int, usr *tgbotapi.User) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var task *Task
	for _, t := range r.tasks {
		if t.ID == taskID {
			task = t
		}
	}

	if task == nil {
		return false, nil
	}

	task.assignedTo = usr
	return true, nil
}

func (r *TasksRepoInMemory) Unassign(taskID int, usr *tgbotapi.User) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task := r.get(taskID)

	if task == nil {
		return false, nil
	}

	task.assignedTo = nil
	return true, nil
}

func (r *TasksRepoInMemory) Delete(id int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, t := range r.tasks {
		if t.ID == id {
			l := len(r.tasks)
			r.tasks[i] = r.tasks[l-1]
			r.tasks = r.tasks[:l-1]
			return true, nil
		}
	}

	return false, nil
}

func (r *TasksRepoInMemory) GetByID(id int) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tasks {
		if t.ID == id {
			return t, nil
		}
	}

	return nil, nil
}

func (r *TasksRepoInMemory) GetByAssigneeID(id int) ([]*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]*Task, 0)
	for _, t := range r.tasks {
		if t.assignedTo != nil && t.assignedTo.ID == id {
			res = append(res, t)
		}
	}

	return res, nil
}

func (r *TasksRepoInMemory) GetByAuthorID(id int) ([]*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]*Task, 0)
	for _, t := range r.tasks {
		if t.createdBy.ID == id {
			res = append(res, t)
		}
	}

	return res, nil
}

func (r *TasksRepoInMemory) get(id int) *Task {
	for _, t := range r.tasks {
		if t.ID == id {
			return t
		}
	}

	return nil
}
