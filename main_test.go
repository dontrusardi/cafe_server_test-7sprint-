package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	request := []string {
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range request {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	request := []struct {
		request string
		status int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range request {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response,req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t,v.message, strings.TrimSpace(response.Body.String()))	
	}
}

//проверяет работу сервера при разных знач count
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	request := []struct {
		city string // город moscow or tula
		count int // передаваемое знач count
		want int // ожидаемое кол-во кафе в ответе
	}{
		{"moscow", 0, 0},
		{"moscow", 1, 1},
		{"moscow", 2, 2},
		{"moscow", 100, len(cafelist["moscow"])},
		{"tula", 0, 0},
		{"tula", 1, 1},
		{"tula", 2, 2},
		{"tula", 100, len(cafelist["tula"])},
	}

	for _, v := range request { 
		url := "/cafe?city=" + v.city + "&count=" + strconv.Itoa(v.count)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)
		handler.ServeHTTP(response, req)
		
		// тут проверяем усепшно ли обработали запрос
		require.Equal(t, http.StatusOK, response.Code,"Wait status 200 OK")

		// тут ответ и  минус лишние пробелы и символы
		responseBody := strings.TrimSpace(response.Body.String())

		var item []string
		if responseBody == "" {
			item = []string{}
		} else {
			item = strings.Split(responseBody, ",")
		}

		// а тут корректный ли результат
		assert.Len(t, item, v.want)
	}
}

//проверяет рез поиска по кафе, по подстроке!
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	request := []struct {
		search string // знач search
		wantCount int // ожидаемое кол-во кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range request {
		url := "/cafe?city=moscow&search=" + v.search
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code, "Wait status 200 OK")
	
		responseBody := strings.TrimSpace(response.Body.String())

		var items []string
		if responseBody == "" {
			items = []string{}
		} else {
			items = strings.Split(responseBody, ",")
		}
		
		checkSubStr := true
		for _, c := range items {
			if !strings.Contains(strings.ToLower(c), strings.ToLower(v.search)) {
				checkSubStr = false
				break
			}
		}

		require.Equal(t, v.wantCount, len(items))
		require.True(t, checkSubStr, v.search)
	}

}