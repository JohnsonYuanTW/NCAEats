package handler

import (
	"fmt"
	"os"

	"path/filepath"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type TemplateHandler struct {
	templates map[string]string
}

func NewTemplateHandler(dir string) (*TemplateHandler, error) {
	templates := make(map[string]string)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// load the file content
		content, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		// get the base name of the file (without extension)
		baseName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		// store it in the map
		templates[baseName] = string(content)
	}

	return &TemplateHandler{templates: templates}, nil
}

func (t *TemplateHandler) GetTemplate(name string) (string, error) {
	template, exists := t.templates[name]
	if !exists {
		return "", fmt.Errorf("template %s does not exist", name)
	}
	return template, nil
}

// Unmarshal JSON template
type unmarshalFunc func([]byte) (interface{}, error)

func (t *TemplateHandler) getComponent(name string, unmarshal unmarshalFunc, data ...interface{}) (interface{}, error) {
	// Get template by name
	template, err := t.GetTemplate(name)
	if err != nil {
		return nil, err
	}

	// Insert data (if any) into the template
	if len(data) != 0 {
		template = fmt.Sprintf(template, data...)
	}

	// Parse JSON to linebot flex container
	component, err := unmarshal([]byte(template))
	if err != nil {
		return nil, err
	}

	return component, nil
}

func unmarshalFlexContainer(data []byte) (interface{}, error) {
	flexContainer, err := linebot.UnmarshalFlexMessageJSON(data)
	return flexContainer, err
}

func unmarshalBoxComponent(data []byte) (interface{}, error) {
	boxComponent := &linebot.BoxComponent{}
	err := boxComponent.UnmarshalJSON(data)
	return *boxComponent, err
}

func (t *TemplateHandler) generateFlexContainer(name string, data ...interface{}) (linebot.FlexContainer, error) {
	flex, err := t.getComponent(name, unmarshalFlexContainer, data...)
	if err != nil {
		return nil, err
	}
	return flex.(linebot.FlexContainer), nil
}

func (t *TemplateHandler) generateBoxComponent(name string, data ...interface{}) (linebot.BoxComponent, error) {
	box, err := t.getComponent(name, unmarshalBoxComponent, data...)
	if err != nil {
		return linebot.BoxComponent{}, err
	}
	return box.(linebot.BoxComponent), nil
}
