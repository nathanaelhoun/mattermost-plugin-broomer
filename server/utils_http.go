package main

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func (p *Plugin) respondAndLogErr(w http.ResponseWriter, code int, err error) {
	http.Error(w, err.Error(), code)
	p.API.LogError(err.Error())
}

func (p *Plugin) respondJSON(w http.ResponseWriter, obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		p.respondAndLogErr(w, http.StatusInternalServerError, errors.WithMessage(err, "failed to marshal response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		p.respondAndLogErr(w, http.StatusInternalServerError, errors.WithMessage(err, "failed to write response"))
		return
	}

	w.WriteHeader(http.StatusOK)
}
