package main

import (
	"fmt"
	"log"

	"github.com/suborbital/reactr/rcap"
)

func main() {
	gqlClient := rcap.NewGraphQLClient()

	resp, err := gqlClient.Do("https://api.rawkode.dev", `{
		allProfiles {
			forename
			surname
		}
	}
	`)

	if err != nil {
		fmt.Println(resp)
		log.Fatal(err)
	}

	fmt.Println(resp.Data)
}
