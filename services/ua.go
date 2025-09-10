package services;

import (
	"fmt"
	"math/rand"
	"time"
	"github.com/go-faker/faker/v4"
)

func GenerateModernChromeUA() string {
	faker.SetRandomSource(rand.NewSource(time.Now().UnixNano()))

	chromeVersions := []string{
		"120.0.6099.109", "121.0.6167.139", "122.0.6261.94", "123.0.6312.105",
		"124.0.6367.78", "125.0.6422.141", "126.0.6478.114", "127.0.6533.88",
		"128.0.6613.113", "129.0.6668.89", "130.0.6723.116",
	}

	windowsVersions := []string{
		"Windows NT 10.0; Win64; x64",
		"Windows NT 11.0; Win64; x64",
	}

	macVersions := []string{
		"Macintosh; Intel Mac OS X 10_15_7",
		"Macintosh; Intel Mac OS X 11_7_10",
		"Macintosh; Intel Mac OS X 12_7_4",
		"Macintosh; Intel Mac OS X 13_6_6",
		"Macintosh; Intel Mac OS X 14_4_1",
	}

	platforms := []string{}
	platforms = append(platforms, windowsVersions...)
	platforms = append(platforms, macVersions...)

	chromeIdx := rand.Intn(len(chromeVersions))
	platformIdx := rand.Intn(len(platforms))

	chromeVersion := chromeVersions[chromeIdx]
	platform := platforms[platformIdx]

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
		platform, chromeVersion)
}