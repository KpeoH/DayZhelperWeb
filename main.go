package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Определение структур данных
type Weapon struct {
	Name   string  `json:"name"`
	Damage float64 `json:"damage"` // Единое значение урона оружия
}

type WallType string

const (
	Wooden WallType = "wooden"
	Metal  WallType = "metal"
	Frame           = "frame"
)

type RaidCalcRequest struct {
	Weapon           string   `form:"weapon"`
	WallType         WallType `form:"wall_type"`
	HealthMultiplier string   `form:"health_multiplier"`
}

type PasswordGenRequest struct {
	Length int `form:"length"`
	Count  int `form:"count"`
}

// Загрузка данных об оружии
var weapons []Weapon

func init() {
	// Загрузка данных об оружии из weapons.json
	data, err := os.ReadFile("data/weapons.json")
	if err != nil {
		log.Fatal("Ошибка загрузки данных об оружии:", err)
	}

	err = json.Unmarshal(data, &weapons)
	if err != nil {
		log.Fatal("Ошибка разбора данных об оружии:", err)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/password", passwordHandler)
	http.HandleFunc("/raid", raidHandler)
	http.HandleFunc("/calculate", calculateHandler)
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/save-passwords", savePasswordsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("DayzHelper2.0 запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
	tmpl.Execute(w, nil)
}

func passwordHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/password.html"))
	tmpl.Execute(w, nil)
}

func raidHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/raid.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Weapons": weapons,
	})
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не допустим", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	req := RaidCalcRequest{
		Weapon:           r.Form.Get("weapon"),
		WallType:         WallType(r.Form.Get("wall_type")),
		HealthMultiplier: r.Form.Get("health_multiplier"),
	}

	// Находим оружие
	var weapon Weapon
	found := false
	for _, w := range weapons {
		if w.Name == req.Weapon {
			weapon = w
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Оружие не найдено", http.StatusBadRequest)
		return
	}

	// Определяем базовое здоровье стены
	var baseWallHealth float64
	switch req.WallType {
	case Wooden:
		baseWallHealth = 16000 // Базовое здоровье деревянной стены
	case Metal:
		baseWallHealth = 21000 // Базовое здоровье металлической стены
	case Frame:
		baseWallHealth = 12000 // Базовое здоровье каркаса
	default:
		http.Error(w, "Неверный тип стены", http.StatusBadRequest)
		return
	}

	// Применяем множитель здоровья
	multiplier, err := strconv.ParseFloat(req.HealthMultiplier, 64)
	if err != nil {
		http.Error(w, "Неверный формат множителя здоровья", http.StatusBadRequest)
		return
	}

	wallHealth := baseWallHealth * multiplier

	// Рассчитываем количество выстрелов
	shotsNeeded := int(wallHealth / weapon.Damage)
	if shotsNeeded <= 0 {
		shotsNeeded = 1
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/result.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Weapon":           req.Weapon,
		"WallType":         req.WallType,
		"HealthMultiplier": req.HealthMultiplier,
		"ShotsNeeded":      shotsNeeded,
	})
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не допустим", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	lengthStr := r.Form.Get("length")
	countStr := r.Form.Get("count")

	length, err := strconv.Atoi(lengthStr)
	if err != nil || (length != 3 && length != 4) {
		http.Error(w, "Длина пароля должна быть 3 или 4", http.StatusBadRequest)
		return
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 100 {
		http.Error(w, "Количество паролей должно быть от 1 до 100", http.StatusBadRequest)
		return
	}

	rand.Seed(time.Now().UnixNano())

	passwords := make([]string, count)
	for i := 0; i < count; i++ {
		password := ""
		for j := 0; j < length; j++ {
			password += strconv.Itoa(rand.Intn(10))
		}
		passwords[i] = password
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/password_result.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Passwords": passwords,
		"Date":      time.Now().Format("2006-01-02"),
	})
}

func savePasswordsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	passwordsStr := r.Form.Get("passwords")
	date := r.Form.Get("date")

	// Обрабатываем срез паролей
	passwords := strings.Split(passwordsStr, " ")
	joinedPasswords := strings.Join(passwords, ", ")

	// Сохраняем пароли в файл
	filename := fmt.Sprintf("passwords_%s.txt", date)
	err := os.WriteFile(filename, []byte(joinedPasswords), 0644)
	if err != nil {
		http.Error(w, "Ошибка сохранения паролей", http.StatusInternalServerError)
		return
	}

	// Отправляем файл пользователю
	http.ServeFile(w, r, filename)
	// Удаляем файл после загрузки
	defer os.Remove(filename)
}
