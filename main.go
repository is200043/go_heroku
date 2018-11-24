package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
)

func main() {
	e := echo.New()

	//env
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	mongoHost := viper.GetString("mongo.host")
	mongoUser := viper.GetString("mongo.user")
	mongoPass := viper.GetString("mongo.pass")
	port := ":" + viper.GetString("port")

	connString := fmt.Sprintf("%v:%v@%v", mongoUser, mongoPass, mongoHost)
	session, err := mgo.Dial(connString)
	if err != nil {
		e.Logger.Fatal(err)
	}

	h := &handler{
		m: session,
	}

	e.Use(middleware.Logger())
	e.GET("/todos", h.list)
	e.GET("/todos/:id", h.view)
	e.PUT("/todos/:id", h.done)
	e.POST("/todos", h.create)
	e.DELETE("/todos/:id", h.delete)
	e.Logger.Fatal(e.Start(port))
}

type todo struct {
	ID    bson.ObjectId `json:"id" bson:"_id"`
	Topic string        `json:"topic" bson:"topic"`
	Done  bool          `json:"done" bson:"done"`
}

type handler struct {
	m *mgo.Session
}

func (h *handler) create(c echo.Context) error {
	session := h.m.Copy()
	defer session.Close()

	var t todo
	if err := c.Bind(&t); err != nil {
		return err
	}

	t.ID = bson.NewObjectId()
	col := session.DB("workshop").C("chonlakorn")
	if err := col.Insert(t); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, t)
}

func (h *handler) list(c echo.Context) error {
	session := h.m.Copy()
	defer session.Close()

	var ts []todo

	col := session.DB("workshop").C("chonlakorn")
	if err := col.Find(nil).All(&ts); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ts)
}

func (h *handler) view(c echo.Context) error {
	session := h.m.Copy()
	defer session.Close()

	id := bson.ObjectIdHex(c.Param("id"))

	var t todo
	col := session.DB("workshop").C("chonlakorn")
	if err := col.FindId(id).One(&t); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, t)
}

func (h *handler) done(c echo.Context) error {
	session := h.m.Copy()
	defer session.Close()

	id := bson.ObjectIdHex(c.Param("id"))

	var t todo
	col := session.DB("workshop").C("chonlakorn")
	if err := col.FindId(id).One(&t); err != nil {
		return nil
	}
	t.Done = true
	if err := col.UpdateId(id, t); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, t)
}

func (h *handler) delete(c echo.Context) error {
	session := h.m.Copy()
	defer session.Close()

	id := bson.ObjectIdHex(c.Param("id"))

	col := session.DB("workshop").C("chonlakorn")
	if err := col.RemoveId(id); err != nil {
		return nil
	}

	return c.JSON(http.StatusOK, echo.Map{"result": "success"})
}
