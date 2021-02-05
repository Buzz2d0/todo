package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails"
)

// Todos struct
type Todos struct {
	homedir  string
	filename string
	runtime  *wails.Runtime
	logger   *wails.CustomLogger
	watcher  *fsnotify.Watcher
}

// NewTodos attempts to create a new Todo list
func NewTodos() (*Todos, error) {

	result := &Todos{}
	return result, nil
}

func (t *Todos) startWatcher() error {
	t.logger.Info("Starting Watcher")
	watcher, err := fsnotify.NewWatcher()
	t.watcher = watcher
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					t.logger.Infof("modified file: %s", event.Name)
					t.runtime.Events.Emit("filemodified")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				t.logger.Error(err.Error())
			}
		}
	}()

	err = watcher.Add(t.filename)
	if err != nil {
		return err
	}
	return nil
}

// GetSrcFilename 数据保存的文件
func (t *Todos) GetSrcFilename() string {
	return t.filename
}

// LoadList 加载 todo list
func (t *Todos) LoadList() (string, error) {
	t.logger.Infof("Loading list from: %s", t.filename)
	bytes, err := ioutil.ReadFile(t.filename)
	if err != nil {
		err = fmt.Errorf("Unable to open list: %s", t.filename)
	}
	return string(bytes), err
}

func (t *Todos) saveListByName(todos string, filename string) error {
	return ioutil.WriteFile(filename, []byte(todos), 0600)
}

// SaveList 持久化 todo list
func (t *Todos) SaveList(todos string) error {
	return t.saveListByName(todos, t.filename)
}

func (t *Todos) setFilename(filename string) error {
	var err error
	// Stop watching the current file and return any error
	err = t.watcher.Remove(t.filename)
	if err != nil {
		return err
	}

	// Set the filename
	t.filename = filename

	// Add the new file to the watcher and return any errors
	err = t.watcher.Add(filename)
	if err != nil {
		return err
	}
	t.logger.Info("Now watching: " + filename)
	return nil
}

// SaveAs 导出 todo list
func (t *Todos) SaveAs(todos string) error {
	filename := t.runtime.Dialog.SelectSaveFile("导出文件", "*.json")
	t.logger.Info("==========Save As: " + filename)
	err := t.saveListByName(todos, filename)
	if err != nil {
		return err
	}
	return t.setFilename(filename)
}

// LoadNewList 导入新的 todo list
func (t *Todos) LoadNewList() {
	filename := t.runtime.Dialog.SelectFile("导入文件", "*.json")
	if len(filename) > 0 {
		t.setFilename(filename)
		t.runtime.Events.Emit("filemodified")
	}
}

func (t *Todos) getHomedir() error {
	var (
		homedir string
		err     error
	)
	if homedir, err = t.runtime.FileSystem.HomeDir(); err != nil {
		return err
	}
	t.homedir = path.Join(homedir, ".todos/")
	t.filename = path.Join(t.homedir, "todos.json")
	return nil
}

func (t *Todos) ensureFileExists() {
	var err error
	_, err = os.Stat(t.homedir)
	if os.IsNotExist(err) {
		os.Mkdir(t.homedir, 0755)
	}
	_, err = os.Stat(t.filename)
	if os.IsNotExist(err) {
		ioutil.WriteFile(t.filename, []byte("[]"), 0600)
	}
}

// WailsInit wails 初始化
func (t *Todos) WailsInit(runtime *wails.Runtime) error {
	t.runtime = runtime
	t.logger = t.runtime.Log.New("Todos")
	t.getHomedir()
	t.ensureFileExists()
	t.runtime.Window.SetTitle("Todo List")
	return t.startWatcher()
}
