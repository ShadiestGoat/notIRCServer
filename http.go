package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/notIRCServer/db"
	"github.com/shadiestgoat/notIRCServer/users"
	"github.com/shadiestgoat/notIRCServer/utils"
	"github.com/shadiestgoat/notIRCServer/ws"
)

type ctxKeys int

const (
	CTX_USER ctxKeys = iota

	HTTP_SESSION_ID = "Session-Id"
)

type HTTPErr struct {
	Err string `json:"error"`
}

type HTTPMsg struct {
	Msg string `json:"message"`
}

func httpWrite(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func httpErr(w http.ResponseWriter, err error) {
	status := 500

	if err, ok := err.(utils.HTTPStatus); ok {
		s := err.GetStatus()
		if s != 0 {
			status = s
		}
	}

	httpWrite(w, status, HTTPErr{Err: err.Error()})
}

type UserResp struct {
	Name  string `json:"name"`
	Color int    `json:"color"`
	// Flags UserFlag `json:"flags"`
}

// type UserFlag int

// const (
// 	UF_CAN_HEAR_WHISPERS UserFlag = 1 << iota
// 	UF_CAN_WHISPER_TO
// )

func Router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			s := strings.Split(auth, " ")

			var u *users.User
			var token string

			if len(s) >= 2 {
				u = users.GetUser(strings.Join(s[:len(s)-1], " "))
				token = s[len(s)-1]
			}

			if u == nil || u.Token != token {
				httpErr(w, utils.ErrNotAuthorized)
			} else {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CTX_USER, u)))
			}
		})
	})

	r.Get("/messages", func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(CTX_USER).(*users.User)

		if !u.HasPerm(users.PERM_READ) {
			httpErr(w, utils.ErrBadPerms)
			return
		}

		msgs := db.GetMessages()

		render.JSON(w, r, db.FilterForUser(msgs, u))
	})

	r.Post(`/messages`, func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(CTX_USER).(*users.User)

		msg := &db.MsgBase{}

		err := json.NewDecoder(r.Body).Decode(msg)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		if msg == nil {
			httpErr(w, utils.HTTPErr{
				Msg:    "Bad msg data",
				Status: 400,
			})
			return
		}

		if msg.To == "" {
			msg.To = "*"
		}

		if msg.To == "*" {
			if !u.HasPerm(users.PERM_WRITE) {
				httpErr(w, utils.ErrBadPerms)
				return
			}
		} else if !u.WriteWhispers[msg.To] {
			httpErr(w, utils.ErrBadPerms)
			return
		}

		m := &db.Msg{
			MsgBase: *msg,
			Author:  u.Name,
		}

		err = db.AddMsg(m)

		if err != nil {
			httpErr(w, err)
			return
		}

		httpWrite(w, 200, m)

		ws.WriteMsg(m, r.Header.Get(HTTP_SESSION_ID))
	})

	r.Delete("/messages", func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(CTX_USER).(*users.User)

		if !u.HasPerm(users.PERM_DELETE) {
			httpErr(w, utils.ErrBadPerms)
			return
		}

		db.DeleteLast()

		httpWrite(w, 200, HTTPMsg{
			Msg: "Msg deleted",
		})
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		// u := r.Context().Value(CTX_USER).(*users.User)

		resp := []*UserResp{}

		all := users.All()

		for n, cu := range all {
			r := &UserResp{
				Name:  n,
				Color: cu.Color,
			}

			resp = append(resp, r)
		}

		httpWrite(w, 200, resp)
	})

	r.Get("/perms", func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(CTX_USER).(*users.User)
		httpWrite(w, 200, u)
	})

	r.HandleFunc(`/ws`, func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(CTX_USER).(*users.User)

		if !u.HasPerm(users.PERM_READ_LIVE) {
			httpErr(w, utils.ErrBadPerms)
			return
		}

		cID := uuid.NewString()

		conn, err := wsUp.Upgrade(w, r, http.Header{
			HTTP_SESSION_ID: {cID},
		})

		if log.ErrorIfErr(err, "upgrading ws") {
			return
		}

		ws.AddConn(u, cID, conn)
	})

	return r
}
