package main

import "strings"

func StationsToArray(stations string) []string {
	stationsArray := []string{}

	for _, iaco := range strings.Split(stations, ",") {
		stationsArray = append(stationsArray, strings.ReplaceAll(iaco, " ", ""))
	}

	return stationsArray
}

func StationsArrayToIDs(stations []string) string {
	returnString := ""

	for _, id := range stations {
		returnString += id + ","
	}

	return returnString[:len(returnString)-1]

}
