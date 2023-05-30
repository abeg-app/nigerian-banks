package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Error struct {
	Message string
}

type BankJSON struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Code string `json:"code"`
	USSD string `json:"ussd"`
}

type Bank struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Code string `json:"code"`
	USSD string `json:"ussd"`
	Logo string `json:"logo"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	port, exist := os.LookupEnv("PORT")

	if !exist {
		log.Fatal("PORT not set in .env")
	}

	host, exist := os.LookupEnv("HOST")

	if !exist {
		log.Fatal("HOST not set in .env")
	}

	bankJson, err := ioutil.ReadFile("./banks.json")

	if err != nil {
		log.Fatal(err)
	}

	var banks []BankJSON

	if err := json.Unmarshal(bankJson, &banks); err != nil {
		log.Fatal(err)
	}

	route := mux.NewRouter()

	route.PathPrefix("/logo/").Handler(http.StripPrefix("/logo/", http.FileServer(http.Dir("./logos"))))

	route.NotFoundHandler = http.HandlerFunc(notFound)

	route.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		code := request.URL.Query().Get("code")
		slug := request.URL.Query().Get("slug")

		if code != "" || slug != "" {
			// Perform an SQL "and" query to find the exact match
			for _, bank := range banks {
				if (code == "" || code == bank.Code) && (slug == "" || slug == bank.Slug) {
					_ = json.NewEncoder(writer).Encode(Bank{
						Name: bank.Name,
						Slug: bank.Slug,
						Code: bank.Code,
						USSD: bank.USSD,
						Logo: host + "/logo/" + getUrl(bank.Slug) + ".png",
					})
					return
				}
			}
			_ = json.NewEncoder(writer).Encode(nil)
			return
		}

		// No code and slug provided, return all banks
		var newBanks []Bank

		for _, bank := range banks {
			newBanks = append(newBanks, Bank{
				Name: bank.Name,
				Slug: bank.Slug,
				Code: bank.Code,
				USSD: bank.USSD,
				Logo: host + "/logo/" + getUrl(bank.Slug) + ".png",
			})
		}

		_ = json.NewEncoder(writer).Encode(newBanks)
	})

	handler := cors.AllowAll().Handler(route)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func notFound(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)

	writer.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(writer).Encode(Error{
		Message: "endpoint not found",
	})
}

func getUrl(slug string) string {
	var files []string

	f, err := os.Open("./logos")

	if err != nil {
		log.Fatal(err)
	}

	fileInfo, err := f.Readdir(0)

	_ = f.Close()

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range fileInfo {
		if file.Name() == ".DS_Store" {
			continue
		}

		files = append(files, file.Name())
	}

	_, found := find(files, slug+".png")

	if found {
		return slug
	}

	return "default-image"
}

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
