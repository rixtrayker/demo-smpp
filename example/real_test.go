package main

import (
	"os"
	"sync"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/session"
)

var providersList []config.Provider

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	
	go testLive(&wg)

	wg.Wait()
}

func testLive(wg *sync.WaitGroup) {
	defer wg.Done()
	
	cfg := config.LoadConfig()
	providersList = cfg.ProvidersConfig
	// sending SMS(s)
	msg := "testing GoLang worker response"
	// load numbers from .env testing zain_number, mobily_number, stc_number
	zain_number := os.Getenv("ZAIN_NUMBER")
	mobily_number := os.Getenv("MOBILY_NUMBER")
	stc_number := os.Getenv("STC_NUMBER") 
	cfg  := config.LoadConfig()

	if(zain_number != "" && ){
		testZain(zain_number,msg)
	}

}

func getProviderCfg(list *[]config.Provider, name string) config.Provider{
	for provider := range *list {
		if provider.name == name {
			return provider
		}
	}
}

func testZain(zain_number, msg string) {
	cfg := getProviderCfg("Zain")
	if cfg == (config.Provider{}) {
		log.error("empty zain config")
		return 
	}

	zain_session, err := session.NewSession(cfg)
	if err != nil {
		log.error(err)
	}
	zain_session.Send(zain_number, msg)
}

// testMobily func
func testMobily(mobily_number, msg string) {
	cfg := getProviderCfg("Mobily")
	if cfg == (config.Provider{}) {
		log.error("empty mobily config")
		return 
	}

	mobily_session, err := session.NewSession(cfg)
	if err != nil {
		log.error(err)
	}
	mobily_session.Send(mobily_number, msg)
}

// testSTC func
func testSTC(stc_number, msg string) {
	cfg := getProviderCfg("STC")
	if cfg == (config.Provider{}) {
		log.error("empty stc config")
		return 
	}

	stc_session, err := session.NewSession(cfg)
	if err != nil {
		log.error(err)
	}
	stc_session.Send(stc_number, msg)
}