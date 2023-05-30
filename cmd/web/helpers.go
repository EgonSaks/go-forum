package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"forum/logger"
)

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, name string, td *templateData) error {
	ts, ok := app.templateCache[name]
	if !ok {
		logger.ErrorLogger.Println("The template does not exist")
		return fmt.Errorf("the template %s does not exist", name)
	}

	buf := new(bytes.Buffer)

	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		logger.ErrorLogger.Printf("Error executing template: %v\n", err)
		return err
	}

	buf.WriteTo(w)
	return nil
}
