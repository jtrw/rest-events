package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"io"
	repository "micro-rest-events/v1/app/backend/repository"
	"net/http"
	"strconv"
	"log"
)

const STATUS_NEW = "new"

type JSON map[string]interface{}

type Handler struct {
	Connection *sql.DB
	EventRepository repository.EventRepositoryInterface
}


func NewHandler(rep repository.EventRepositoryInterface) Handler {
	return Handler{EventRepository: rep}
}
//
// func (h Handler) NewEventRepository() repository.EventRepositoryInterface {
//     return repository.NewEventRepository(h.Connection)
// }

func (h Handler) OnCreateEvent(w http.ResponseWriter, r *http.Request) {
	var requestData JSON

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		render.Status(r, http.StatusNotFound)
		return
	}

	err = json.Unmarshal(b, &requestData)

	if err != nil {
		log.Println("[ERROR] Error while decoding the data", err.Error())
		render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": "Error while decoding the data"})
		return
	}

	if requestData["type"] == nil || requestData["user_id"] == nil {
	    render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": "Not found type or user_id"})
        return
	}

	uuid := uuid.New().String()

	userId := int(requestData["user_id"].(float64))
	rec := repository.Event{
		Uuid:   uuid,
		UserId: userId,
		Status: STATUS_NEW,
		Type: requestData["type"].(string),
	}

	eventRepository := h.EventRepository
	err = eventRepository.Create(rec)

	if err != nil {
	    log.Println(err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": err})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, JSON{"status": "ok", "uuid": uuid})
}

func (h Handler) OnGetEventsByUserId(w http.ResponseWriter, r *http.Request) {
	userId, _ := strconv.Atoi(chi.URLParam(r, "id"))

    eventRepository := h.EventRepository
	row, err := eventRepository.GetByUserId(userId)

	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, JSON{"status": "not_found", "message": "Not Found"})
		return
	}

	render.JSON(w, r, JSON{"status": "ok", "data": row})
}

func (h Handler) OnChangeEvent(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")

	var requestData JSON

	b, err := io.ReadAll(r.Body)
	if err != nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": "Error read the data"})
        return
	}

	err = json.Unmarshal(b, &requestData)

	if err != nil {
	    render.Status(r, http.StatusBadRequest)
	    render.JSON(w, r, JSON{"status": "error", "message": "Error while decoding the data"})
	    return
	}

	if requestData["status"] == nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": "Status is required"})
        return
    }

	var message string = ""
	if requestData["message"] != nil {
		message = requestData["message"].(string)
	}

	rec := repository.Event{
		Status:  requestData["status"].(string),
		Message: message,
	}

	eventRepository := h.EventRepository
	count, err := eventRepository.ChangeStatus(uuid, rec)

    if err != nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": err})
    }

	if count == 0 {
		// Check if the row not exists
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, JSON{"status": "error", "message": "Not Found"})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, JSON{"status": "ok", "uuid": uuid})
}

func (h Handler) OnSetSeen(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")

    eventRepository := h.EventRepository
	count, err := eventRepository.ChangeIsSeen(uuid)

    if err != nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": err})
    }

	if count == 0 {
		// Check if the row not exists
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, JSON{"status": "error", "message": "Not Found"})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, JSON{"status": "ok", "uuid": uuid})
}
