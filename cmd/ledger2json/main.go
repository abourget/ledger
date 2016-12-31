package ledger

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/abourget/ledger/parse"
)

func main() {
	cnt, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln("Error reading input:", err)
	}

	err = parse.New("stdin", string(cnt)).Parse()
	if err != nil {
		log.Fatalln("Parsing error:", err)
	}


}
