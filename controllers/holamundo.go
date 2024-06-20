package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
)

var servers = []server{
	{puerto: 5050, trafico: 0},
	{puerto: 7070, trafico: 0},
	//{puerto: 6060, trafico: 0},
	//{puerto: 9090, trafico: 0},
}

type server struct {
	puerto  int
	trafico int
}

var mutex sync.Mutex

func ManejarRequest(e echo.Context) error {
	serversNoDisponibles := make([]int, 0, 16)

	res, err := selectServer(serversNoDisponibles, e)
	if err != nil {
		return e.String(503, "el servicio no está disponible")
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = e.String(200, string(resBody))
	if err != nil {
		return err
	}
	return nil
}

func selectServer(serversNoDisponibles []int, e echo.Context) (http.Response, error) {

	menorTrafico := servers[0].trafico
	puertoCorrespondiente := servers[0].puerto
	indexMenorTrafico := 0

	for i := 1; i < len(servers); i++ {
		mutex.Lock()
		if len(servers) == len(serversNoDisponibles) {
			err := fmt.Errorf("no hay servers disponibles")
			var response http.Response
			fmt.Printf("Aca retorna 1")
			return response, err
		} else if contiene(servers[i].puerto, serversNoDisponibles) {
			continue
		} else if servers[i].trafico < menorTrafico {
			menorTrafico = servers[i].trafico
			puertoCorrespondiente = servers[i].puerto
			indexMenorTrafico = i
		}
		mutex.Unlock()
	}
	mutex.Lock()
	servers[indexMenorTrafico].trafico += 1
	mutex.Unlock()
	defer func() {
		mutex.Lock()
		servers[indexMenorTrafico].trafico -= 1
		mutex.Unlock()
	}()
	fmt.Println(puertoCorrespondiente, "este es el puerto seleccionado")
	res, err := redirigirRequest(puertoCorrespondiente, e, serversNoDisponibles)
	return *res, err
}

func redirigirRequest(port int, e echo.Context, serversNoDisponibles []int) (*http.Response, error) {

	path := e.Param("*")
	method := e.Request().Method
	bodyStream := e.Request().Body
	headers := e.Request().Header
	fmt.Println(path)

	requestURL := fmt.Sprintf("http://localhost:%s/%s", strconv.Itoa(port), path)
	req, err := http.NewRequest(method, requestURL, bodyStream)
	for headerName, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
	}
	*res, err = manejarRespuesta(port, *res, e, err, serversNoDisponibles)
	if err != nil {
		fmt.Println("la concha de tu vieja")
	}

	return res, err
}

func manejarRespuesta(port int, res http.Response, e echo.Context, err error, serversNoDisponibles []int) (http.Response, error) {
	if res.StatusCode > 499 && res.StatusCode < 600 {
		serversNoDisponibles = append(serversNoDisponibles, port)
		newRes, err := selectServer(serversNoDisponibles, e)
		if err != nil {
			fmt.Printf("Aca retorna 3")
			return newRes, err
		}
		res = newRes
	}
	for _, server := range serversNoDisponibles {
		fmt.Println(server, "este es un server no disponible")
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	return res, err
}

func contiene(server int, serversNoDisponibles []int) bool {

	for _, puerto := range serversNoDisponibles {
		if puerto == server {
			return true
		}
	}
	return false
}
