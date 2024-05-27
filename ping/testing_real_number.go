package ping

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/session"
)

var providersList []config.Provider
var ctx = context.Background()


func TestLive(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Testing Real Numbers")
	cfg := config.LoadConfig()

	providersList = cfg.ProvidersConfig
	// sending SMS(s)
	msg := "testing GoLang worker response"
	// load numbers from .env testing zain_number, mobily_number, stc_number
	zain_number := os.Getenv("ZAIN_NUMBER")
	mobily_number := os.Getenv("MOBILY_NUMBER")
	stc_number := os.Getenv("STC_NUMBER") 

	if(zain_number != "") {
		fmt.Println("Testing Zain Number")
		wg.Add(1)
		go func() {
			defer wg.Done()
			testZain(zain_number,msg)
		}()
	}
	if(mobily_number != "") {
		fmt.Println("Testing Mobily Number")
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMobily(mobily_number,msg)
		}()
	}
	if(stc_number != "") {
		fmt.Println("Testing STC Number")
		wg.Add(1)
		go func() {
			defer wg.Done()
			testSTC(stc_number,msg)
		}()
	}

}


func getProviderCfg(name string) config.Provider {
	for _, provider := range providersList {
		if provider.Name == name {
			return provider
		}
	}
	return config.Provider{}
}

func testZain(zain_number, msg string) {
	cfg := getProviderCfg("Zain")
	if cfg == (config.Provider{}) {
		fmt.Println("empty zain config")
		return 
	}

	zain_session, err := session.NewSession(ctx, cfg, nil)
	if err != nil {
		fmt.Println(err)
	}
	zain_session.Send(zain_number, msg)
	// dump session message ids
	time.Sleep(30000 * time.Millisecond)

	fmt.Println("MSGIDS: ",zain_session.MessageIDs)
}

// testMobily func
func testMobily(mobily_number, msg string) {
	cfg := getProviderCfg("Mobily")
	if cfg == (config.Provider{}) {
		fmt.Println("empty mobily config")
		return 
	}

	mobily_session, err := session.NewSession(ctx, cfg, nil)
	if err != nil {
		fmt.Println(err)
	}
	mobily_session.Send(mobily_number, msg)
	// dump session message ids
	time.Sleep(30000 * time.Millisecond)

	fmt.Println("MSGIDS: ",mobily_session.MessageIDs)
}

// testSTC func
func testSTC(stc_number, msg string) {
	cfg := getProviderCfg("STC")
	if cfg == (config.Provider{}) {
		fmt.Println("empty stc config")
		return 
	}

	stc_session, err := session.NewSession(ctx, cfg, nil)
	if err != nil {
		fmt.Println(err)
	}
	stc_session.Send(stc_number, msg)
	// dump session message ids
	time.Sleep(30000 * time.Microsecond)
	fmt.Println("MSGIDS: ",stc_session.MessageIDs)
}