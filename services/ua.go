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

func GetRandomAndroidUA() string {
	user_agents := []string{
		"Mozilla/5.0 (Linux; Android 5.0.2; Andromax C46B2G Build/LRX22G) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/37.0.0.0 Mobile Safari/537.36 [FB_IAB/FB4A;FBAV/60.0.0.16.76;]",
		"[FBAN/FB4A;FBAV/35.0.0.48.273;FBDM/{density=1.33125,width=800,height=1205};FBLC/en_US;FBCR/;FBPN/com.facebook.katana;FBDV/Nexus 7;FBSV/4.1.1;FBBK/0;]",
		"Mozilla/5.0 (Linux; Android 5.1.1; SM-N9208 Build/LMY47X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.81 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; U; Android 5.0; en-US; ASUS_Z008 Build/LRX21V) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 UCBrowser/10.8.0.718 U3/0.8.0 Mobile Safari/534.30",
		"Mozilla/5.0 (Linux; U; Android 5.1; en-US; E5563 Build/29.1.B.0.101) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 UCBrowser/10.10.0.796 U3/0.8.0 Mobile Safari/534.30",
		"Mozilla/5.0 (Linux; U; Android 4.4.2; en-us; Celkon A406 Build/MocorDroid2.3.5) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	}
	return user_agents[rand.Intn(len(user_agents))]
}