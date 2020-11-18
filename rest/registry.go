package rest

import (
	"log"
	"time"

	"golang.org/x/sys/windows/registry"
)

func verifyDSNValues(DSN string, DB string, driver string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\ODBC\ODBC.INI\`+DSN, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\` + DSN)
		log.Fatal(err)
	}
	err = k.SetStringValue("Driver", driver)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\` + DSN)
		log.Fatal(err)
	}
	err = k.SetStringValue("Encryption", "")
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\` + DSN)
		log.Fatal(err)
	}
	err = k.SetStringValue("IntegrityCheck", "0")
	if err != nil {
		log.Fatal(err)
	}
	err = k.SetStringValue("PWDXX", "")
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\` + DSN)
		log.Fatal(err)
	}
	err = k.SetStringValue("RepFic", DB)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\` + DSN)
		log.Fatal(err)
	}
	t := time.Now()
	err = k.SetStringValue("Updated", t.String())
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\` + DSN)
		log.Fatal(err)
	}

}

func verifyDSN(DSN string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\ODBC\ODBC.INI\ODBC Data Sources`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\ODBC Data Sources`)
		log.Fatal(err)
	}
	defer k.Close()
	val, _, err := k.GetStringValue(DSN)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\ODBC Data Sources`)
		log.Fatal(err)
	}
	if val == "HFSQL" {
		log.Println(DSN + " NAME OK")
		return true
	}
	err = k.SetStringValue(DSN, "HFSQL")
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\ODBC Data Sources`)
		log.Fatal(err)
	}
	log.Println(DSN + " ADDED SUCESSFULLY")
	return true
}

func setKeyValueDSN(DSN string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\ODBC\ODBC.INI\ODBC Data Sources`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\ODBC Data Sources`)
		log.Fatal(err)
	}
	defer k.Close()
	err = k.SetStringValue(DSN, "HFSQL")
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\ODBC Data Sources`)
		log.Fatal(err)
	}
	log.Println(DSN + " ADDED SUCESSFULLY")
	return true
}

func UpdateDSN(DSN string, DB string, driver string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\ODBC\ODBC.INI`, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI`)
		log.Fatal(err)
	}
	defer k.Close()

	names, err := k.ReadSubKeyNames(-1)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI`)
		log.Fatal(err)
	}
	for _, name := range names {
		if name == DSN {
			if verifyDSN(DSN) {
				verifyDSNValues(DSN, DB, driver)
			}
			return true
		}
	}
	k2, _, err := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\ODBC\ODBC.INI\`+DSN, registry.ALL_ACCESS)
	if err != nil {
		log.Println(`Write registry error: .CURRENT_USER\SOFTWARE\ODBC\ODBC.INI\` + DSN)
		log.Fatal(err)
	}
	k2.Close()
	if setKeyValueDSN(DSN) {
		verifyDSNValues(DSN, DB, driver)
	}
	return false
}
