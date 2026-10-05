package http

import (
    "bytes"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "io/fs"
    "net/http"
    "path/filepath"
    "strings"
    "time"

    "github.com/umam-src/buku-catatan/internal/db"
    "github.com/umam-src/buku-catatan/internal/ui"
)

type Server struct { db *db.DB }

func New(d *db.DB) http.Handler {
    s := &Server{db: d}
    mux := http.NewServeMux()
    mux.HandleFunc("GET /api/notebooks", s.listNotebooks)
    mux.HandleFunc("POST /api/notebooks", s.createNotebook)
    mux.HandleFunc("GET /api/notebooks/{id}/notes", s.listNotes)
    mux.HandleFunc("POST /api/notebooks/{id}/notes", s.createNote)
    mux.HandleFunc("PUT /api/notes/{id}", s.updateNote)
    mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status":"ok"}) })

    assets, _ := fs.Sub(ui.Assets, "dist")
    fileServer := http.FileServer(http.FS(assets))
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/api/") { http.NotFound(w, r); return }
        p := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
        if p != "." && p != "" {
            if f, err := assets.Open(p); err == nil { f.Close(); fileServer.ServeHTTP(w, r); return }
        }
        index, err := fs.ReadFile(assets, "index.html")
        if err != nil { http.Error(w, "antarmuka belum dibangun", 500); return }
        http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(index))
    })
    return withHeaders(mux)
}

func (s *Server) listNotebooks(w http.ResponseWriter, r *http.Request) {
    rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,description,archived,created_at,updated_at FROM notebooks WHERE archived=0 ORDER BY updated_at DESC`)
    if err != nil { serverError(w, err); return }; defer rows.Close()
    out := []Notebook{}
    for rows.Next() { var n Notebook; var archived int; if err := rows.Scan(&n.ID,&n.Title,&n.Description,&archived,&n.CreatedAt,&n.UpdatedAt); err != nil { serverError(w,err); return }; n.Archived=archived != 0; out=append(out,n) }
    writeJSON(w,200,out)
}

func (s *Server) createNotebook(w http.ResponseWriter, r *http.Request) {
    var in struct { Title string `json:"title"`; Description string `json:"description"` }
    if !decode(w,r,&in) { return }; in.Title = strings.TrimSpace(in.Title); if in.Title == "" { http.Error(w,"Judul wajib diisi",400); return }
    now:=time.Now().UTC().Format(time.RFC3339Nano); id:=newID()
    _,err:=s.db.ExecContext(r.Context(),`INSERT INTO notebooks(id,title,description,created_at,updated_at) VALUES(?,?,?,?,?)`,id,in.Title,in.Description,now,now)
    if err!=nil { serverError(w,err); return }
    writeJSON(w,201,Notebook{ID:id,Title:in.Title,Description:in.Description,CreatedAt:now,UpdatedAt:now})
}

func (s *Server) listNotes(w http.ResponseWriter,r *http.Request){
    rows,err:=s.db.QueryContext(r.Context(),`SELECT id,title,content,note_type,created_at,updated_at FROM notes WHERE notebook_id=? ORDER BY updated_at DESC`,r.PathValue("id")); if err!=nil {serverError(w,err);return}; defer rows.Close()
    out:=[]Note{}; for rows.Next(){var n Note;if err:=rows.Scan(&n.ID,&n.Title,&n.Content,&n.NoteType,&n.CreatedAt,&n.UpdatedAt);err!=nil{serverError(w,err);return};out=append(out,n)};writeJSON(w,200,out)
}

func (s *Server) createNote(w http.ResponseWriter,r *http.Request){
    var in struct{Title string `json:"title"`;Content string `json:"content"`};if !decode(w,r,&in){return};id:=newID();now:=time.Now().UTC().Format(time.RFC3339Nano);nb:=r.PathValue("id")
    _,err:=s.db.ExecContext(r.Context(),`INSERT INTO notes(id,notebook_id,title,content,created_at,updated_at) VALUES(?,?,?,?,?,?)`,id,nb,in.Title,in.Content,now,now);if err!=nil{serverError(w,err);return};writeJSON(w,201,Note{ID:id,Title:in.Title,Content:in.Content,NoteType:"human",CreatedAt:now,UpdatedAt:now})
}

func (s *Server) updateNote(w http.ResponseWriter,r *http.Request){
    var in struct{Title string `json:"title"`;Content string `json:"content"`};if !decode(w,r,&in){return};now:=time.Now().UTC().Format(time.RFC3339Nano);res,err:=s.db.ExecContext(r.Context(),`UPDATE notes SET title=?,content=?,updated_at=? WHERE id=?`,in.Title,in.Content,now,r.PathValue("id"));if err!=nil{serverError(w,err);return};n,_:=res.RowsAffected();if n==0{http.NotFound(w,r);return};writeJSON(w,200,map[string]string{"updated_at":now})
}

type Notebook struct{ID string `json:"id"`;Title string `json:"title"`;Description string `json:"description"`;Archived bool `json:"archived"`;CreatedAt string `json:"created_at"`;UpdatedAt string `json:"updated_at"`}
type Note struct{ID string `json:"id"`;Title string `json:"title"`;Content string `json:"content"`;NoteType string `json:"note_type"`;CreatedAt string `json:"created_at"`;UpdatedAt string `json:"updated_at"`}

func newID() string{b:=make([]byte,16);_,_=rand.Read(b);return hex.EncodeToString(b)}
func decode(w http.ResponseWriter,r *http.Request,v any)bool{if err:=json.NewDecoder(r.Body).Decode(v);err!=nil{http.Error(w,"JSON tidak valid",400);return false};return true}
func writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_=json.NewEncoder(w).Encode(v)}
func serverError(w http.ResponseWriter,_ error){http.Error(w,"kesalahan internal",500)}
func withHeaders(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("Referrer-Policy","no-referrer");next.ServeHTTP(w,r)})}
