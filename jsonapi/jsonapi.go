package jsonapi

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mailinglist/mdb"
	"net/http"
)

func setJsonHeader(w http.ResponseWriter)  {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromJson[T any](body io.Reader,target T)  {
  buf := bufio.NewReader(body)
  decoder := json.NewDecoder(buf)
  decoder.Decode(&target)
}

func returnJson[T any](w http.ResponseWriter,withData func()(T,error)){
  setJsonHeader(w)
  data,serverErr := withData()
  if serverErr != nil {
    log.Println(serverErr)
    w.WriteHeader(500)
    serverErrJson, err := json.Marshal(&serverErr)
    if err != nil {
      log.Println(err)
      return
    }
    w.Write(serverErrJson)
    return
  }
  w.WriteHeader(200)
  dataJson,err := json.Marshal(&data)
  if err != nil {
    log.Println(err)
    w.WriteHeader(500)
      return
  }
  w.Write(dataJson)
}

func returnErr(w http.ResponseWriter, err error,code int)  {
  returnJson(w , func() (interface{}, error) {
    errorMessage := struct {
      Err string
    }{
      Err: err.Error(),
    }
    w.WriteHeader(code)
    return errorMessage,nil
  })
}


func CreateEmail(db *sql.DB) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
      return
    }
    entry := mdb.EmailEntry{}
    fromJson(r.Body, &entry)
    if err := mdb.CreateEmail(db, entry.Email);err != nil {
      returnErr(w, err, 400)
      return
    }
    returnJson(w,func() (interface{}, error){
      log.Printf("JSON CreateEmail: %v\n",entry.Email)
      return mdb.GetEmail(db, entry.Email)
    })
  })

}
func GetEmail(db *sql.DB) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      return
    }
    entry := mdb.EmailEntry{}
    fromJson(r.Body, &entry)
    returnJson(w,func() (interface{}, error){
      log.Printf("JSON GetEmail: %v\n",entry.Email)
      return mdb.GetEmail(db, entry.Email)
    })
  })

}
func DeleteEmail(db *sql.DB) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "DELETE" {
      return
    }
    entry := mdb.EmailEntry{}
    fromJson(r.Body, &entry)
    if err := mdb.DeleteEmail(db, entry.Email);err != nil {
      returnErr(w, err, 400)
      return
    }
    returnJson(w,func() (interface{}, error){
      log.Printf("JSON DeleteEmail: %v\n",entry.Email)
      return mdb.GetEmail(db, entry.Email)
    })
  })

}

func UpdateEmail(db *sql.DB) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "PUT" {
      return
    }
    entry := mdb.EmailEntry{}
    fromJson(r.Body, &entry)
    if err := mdb.UpdateEmail(db, entry);err != nil {
      returnErr(w, err, 400)
      return
    }
    returnJson(w,func() (interface{}, error){
      log.Printf("JSON UpdateEmail: %v\n",entry.Email)
      return mdb.GetEmail(db, entry.Email)
    })
  })

}

func GetEmailBatch(db *sql.DB) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      return
    }
    queryOptions := mdb.GetEmailBatchQueryParams{}
    fromJson(r.Body, &queryOptions)

    if queryOptions.Count <= 0 || queryOptions.Page <= 0 {
      returnErr(w, errors.New("Page and Count fields are required and must be > 0"), 406)
      return
    }
    returnJson(w , func() (interface{}, error) {
      log.Printf("JSON GetEmailBatch: %v\n", queryOptions)
      return mdb.GetEmailBatch(db, queryOptions)
    })
    
  })

}
func Serve(db *sql.DB,bind string)  {
  http.Handle("/email/create", CreateEmail(db))
  http.Handle("/email/get", GetEmail(db))
  http.Handle("/email/get_batch", GetEmailBatch(db))
  http.Handle("/email/update", UpdateEmail(db))
  http.Handle("/email/delete", DeleteEmail(db))
  err := http.ListenAndServe(bind, nil)
  if err != nil {
    log.Fatalf("JSON server error: %v",err)
    
  }
  
}