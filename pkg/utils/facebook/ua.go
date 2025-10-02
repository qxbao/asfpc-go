package facebook;

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
func GetRandomIOSUA() string {
	user_agents := []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/526.0.0.30.105;FBBV/796016684;FBDV/iPhone11,6;FBMD/iPhone;FBSN/iOS;FBSV/18.6.2;FBSS/3;FBCR/;FBID/phone;FBLC/en_GB;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/526.0.0.30.105;FBBV/796016684;FBDV/iPhone11,6;FBMD/iPhone;FBSN/iOS;FBSV/18.6.2;FBSS/3;FBCR/;FBID/phone;FBLC/en_GB;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/525.0.0.24.106;FBBV/791202578;FBDV/iPhone12,1;FBMD/iPhone;FBSN/iOS;FBSV/18.6.2;FBSS/2;FBCR/;FBID/phone;FBLC/en_US;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 13_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBDV/iPhone11,2;FBMD/iPhone;FBSN/iOS;FBSV/13.3.1;FBSS/3;FBID/phone;FBLC/cs_CZ;FBOP/5;FBCR/]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/519.0.0.52.111;FBBV/774183298;FBDV/iPhone12,1;FBMD/iPhone;FBSN/iOS;FBSV/18.5;FBSS/2;FBCR/;FBID/phone;FBLC/en_US;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_7_11 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/524.0.0.38.107;FBBV/789599495;FBDV/iPhone10,3;FBMD/iPhone;FBSN/iOS;FBSV/16.7.11;FBSS/3;FBCR/;FBID/phone;FBLC/en_US;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/524.0.0.38.107;FBBV/789599495;FBDV/iPhone12,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6.2;FBSS/3;FBCR/;FBID/phone;FBLC/hr_HR;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/523.0.0.32.108;FBBV/786592955;FBDV/iPhone17,1;FBMD/iPhone;FBSN/iOS;FBSV/18.6.2;FBSS/3;FBCR/;FBID/phone;FBLC/en_GB;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_7_11 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/523.0.0.32.108;FBBV/786592955;FBDV/iPhone10,5;FBMD/iPhone;FBSN/iOS;FBSV/16.7.11;FBSS/3;FBCR/;FBID/phone;FBLC/en_US;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 Instagram 243.1.0.14.111 (iPhone14,5; iOS 15_5; en_SG; en-AU; scale=3.00; 1170x2532; 382468104)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_8_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/18H107 [FBAN/FBIOS;FBAV/357.0.0.25.109;FBBV/355418224;FBDV/iPhone12,5;FBMD/iPhone;FBSN/iOS;FBSV/14.8.1;FBSS/3;FBID/phone;FBLC/ko_KR;FBOP/5;FBRV/357723095]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G100 [FBAN/FBIOS;FBAV/528.0.0.25.76;FBBV/783401691;FBDV/iPhone15,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6.2;FBSS/3;FBID/phone;FBLC/en_GB;FBOP/5;FBRV/786844321]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22D72 [FBAN/FBIOS;FBAV/527.0.0.42.97;FBBV/780115200;FBDV/iPhone14,3;FBMD/iPhone;FBSN/iOS;FBSV/18.3.1;FBSS/3;FBID/phone;FBLC/en_GB;FBOP/5;FBRV/784433750;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G100 [FBAN/FBIOS;FBAV/528.0.0.25.76;FBBV/783401691;FBDV/iPhone15,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6.2;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/786844321]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G100 Instagram 390.0.0.28.85 (iPhone13,3; iOS 18_6_2; en_US; en; scale=3.00; 1170x2532; IABMV/1; 765313520)",
		"Mozilla/5.0 (iPad; CPU OS 18_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G100 [FBAN/FBIOS;FBAV/527.0.0.42.97;FBBV/780115200;FBDV/iPad13,18;FBMD/iPad;FBSN/iPadOS;FBSV/18.6.2;FBSS/2;FBID/tablet;FBLC/en_US;FBOP/5;FBRV/0]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22D72 Instagram 379.0.0.26.81 (iPhone17,3; iOS 18_3_1; en_US; en; scale=3.00; 1179x2556; 729569027; IABMV/1)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22F76 [FBAN/FBIOS;FBAV/523.0.0.38.95;FBBV/766554692;FBDV/iPhone14,5;FBMD/iPhone;FBSN/iOS;FBSV/18.5;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/771907368;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G90 [FBAN/FBIOS;FBAV/526.0.0.61.97;FBBV/776821927;FBDV/iPhone17,1;FBMD/iPhone;FBSN/iOS;FBSV/18.6.1;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/779405033;IABMV/1]",
		"Mozilla/5.0 (iPad; CPU OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/526.0.0.61.97;FBBV/776821927;FBDV/iPad15,7;FBMD/iPad;FBSN/iPadOS;FBSV/18.6;FBSS/2;FBID/tablet;FBLC/en_US;FBOP/5;FBRV/0]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/526.0.0.61.97;FBBV/776821927;FBDV/iPhone17,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778752024;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22F76 [FBAN/FBIOS;FBAV/520.0.0.38.101;FBBV/756351453;FBDV/iPhone14,7;FBMD/iPhone;FBSN/iOS;FBSV/18.5;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/761622528;IABMV/1]",
		"Mozilla/5.0 (iPad; CPU OS 16_7_11 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/20H360 [FBAN/FBIOS;FBAV/526.0.0.61.97;FBBV/776821927;FBDV/iPad6,12;FBMD/iPad;FBSN/iPadOS;FBSV/16.7.11;FBSS/2;FBID/tablet;FBLC/en_US;FBOP/5;FBRV/0]", 
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone13,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778429392;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22C161 [FBAN/FBIOS;FBDV/iPhone15,5;FBMD/iPhone;FBSN/iOS;FBSV/18.2.1;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22C161 [FBAN/FBIOS;FBAV/501.0.0.49.107;FBBV/699723644;FBDV/iPhone15,5;FBMD/iPhone;FBSN/iOS;FBSV/18.2.1;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/703296132;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/526.0.0.61.97;FBBV/776821927;FBDV/iPhone15,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778736737;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone16,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778334391;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone14,5;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778429392;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22F76 [FBAN/FBIOS;FBAV/526.0.0.61.97;FBBV/776821927;FBDV/iPhone17,3;FBMD/iPhone;FBSN/iOS;FBSV/18.5;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/0;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone15,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778429392;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone17,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778357521;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone17,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778283531;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone13,4;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778334391;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone15,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_GB;FBOP/5;FBRV/778208716;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone13,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778283531;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone16,1;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/777488059;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone17,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778208716;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone15,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778357521;IABMV/1]",
		"Mozilla/5.0 (iPad; CPU OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPad8,5;FBMD/iPad;FBSN/iPadOS;FBSV/18.6;FBSS/2;FBID/tablet;FBLC/en_US;FBOP/5;FBRV/778334391]", 
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone17,3;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778283531;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone17,2;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/777722204;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone14,5;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778247232;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone15,5;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778208716;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/526.0.0.61.97;FBBV/776821927;FBDV/iPhone16,1;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/0;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22G86 [FBAN/FBIOS;FBAV/525.0.0.53.107;FBBV/774177433;FBDV/iPhone15,4;FBMD/iPhone;FBSN/iOS;FBSV/18.6;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/778247232;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22F72 Instagram 398.0.0.13.110 (iPhone13,3; iOS 18_4; en_US; en; scale=3.00; 1170x2532; 714795751; IABMV/1)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22F76 Instagram 378.2.0.37.76 (iPhone14,5; iOS 18_5; en_GB; en-GB; scale=3.00; 1170x2532; 729008693; IABMV/1)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_6_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/517.0.0.38.109;FBBV/766555132;FBDV/iPhone15,5;FBMD/iPhone;FBSN/iOS;FBSV/17.6.1;FBSS/3;FBCR/;FBID/phone;FBLC/en_GB;FBOP/80]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22F76 [FBAN/FBIOS;FBAV/521.0.0.38.98;FBBV/760232234;FBDV/iPhone15,4;FBMD/iPhone;FBSN/iOS;FBSV/18.5;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/764572167;IABMV/1]",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/510.0.0.47.116;FBBV/743276974;FBDV/iPhone16,1;FBMD/iPhone;FBSN/iOS;FBSV/18.5;FBSS/3;FBCR/;FBID/phone;FBLC/en_US;FBOP/80]",
	}
	return user_agents[rand.Intn(len(user_agents))]
}