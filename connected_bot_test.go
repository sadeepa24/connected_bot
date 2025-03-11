package connected_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	connected "github.com/sadeepa24/connected_bot"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/watchman"
	"go.uber.org/zap"
)

var ctx = context.Background()

var zLogger, _ = zap.NewDevelopment()

func TestConnectedBot(t *testing.T) {
	newoption := connected.Botoptions{
		Watchman:   &watchman.Watchmanconfig{},
		Dbpath:     "./newtest.db",
		Ctx:        ctx,
		Bottoken:   "",
		Botmainurl: "https://api.telegram.org/bot",
		Metadata: &controller.MetadataConf{
			GroupID:           -1002325676823,
			ChannelID:         -1002400437670,
			Maxconfigcount:    10,
			LoginLimit:        1,
			BandwidthAvelable: "4000GB",
		},
		Logger:     zLogger,
	}

	Bot, _ := connected.New(newoption)

	if err := Bot.Start(); err != nil {
		zLogger.Fatal(err.Error())
	}
	time.Sleep(15 * time.Minute)
	//sendupdateds()
}

func sendupdateds() {
	GroupID := -1001234567890

	commandcreate := `
	{
	  "update_id": 1234567892,
	  "message": {
	    "message_id": 1,
	    "from": {
	      "id": 6754,
	      "is_bot": false,
	      "first_name": "John",
	      "username": "john_doe",
	      "language_code": "en"
	    },
	    "chat": {
	      "id": 6754,
	      "first_name": "John",
	      "username": "john_doe",
	      "type": "private"
	    },
	    "date": 1609459200,
	    "text": "/create",
	    "entities": [
	      {
	        "offset": 0,
	        "length": 7,
	        "type": "bot_command"
	      }
	    ]
	  }
	}
	`
	commandstart := `
	
		{
	  "update_id": 1234567892,
	  "message": {
	    "message_id": 1,
	    "from": {
	      "id": 6754,
	      "is_bot": false,
	      "first_name": "John",
	      "username": "john_doe",
	      "language_code": "en"
	    },
	    "chat": {
	      "id": 6754,
	      "first_name": "John",
	      "username": "john_doe",
	      "type": "private"
	    },
	    "date": 1609459200,
	    "text": "/start",
	    "entities": [
	      {
	        "offset": 0,
	        "length": 6,
	        "type": "bot_command"
	      }
	    ]
	  }
	}
	`

	commandhelp := `
		{
	  "update_id": 1234567892,
	  "message": {
	    "message_id": 1,
	    "from": {
	      "id": 6754,
	      "is_bot": false,
	      "first_name": "John",
	      "username": "john_doe",
	      "language_code": "en"
	    },
	    "chat": {
	      "id": 6754,
	      "first_name": "John",
	      "username": "john_doe",
	      "type": "private"
	    },
	    "date": 1609459200,
	    "text": "/help",
	    "entities": [
	      {
	        "offset": 0,
	        "length": 5,
	        "type": "bot_command"
	      }
	    ]
	  }
	}
	
	`
	commandsendgift := `
		{
	  "update_id": 1234567892,
	  "message": {
	    "message_id": 1,
	    "from": {
	      "id": 6754,
	      "is_bot": false,
	      "first_name": "John",
	      "username": "john_doe",
	      "language_code": "en"
	    },
	    "chat": {
	      "id": 6754,
	      "first_name": "John",
	      "username": "john_doe",
	      "type": "private"
	    },
	    "date": 1609459200,
	    "text": "/sendgift",
	    "entities": [
	      {
	        "offset": 0,
	        "length": 9,
	        "type": "bot_command"
	      }
	    ]
	  }
	}
	
	`

	userid := rand.Int63()
	fmt.Println(strconv.Itoa(-1001234567890))
	newuserjoin := `
		{
  		"update_id": 123456790,
  		"message": {
  		  "message_id": 2,
  		  "from": {
  		    "id": ` + strconv.Itoa(int(userid)) + `,
  		    "is_bot": false,
  		    "first_name": "Newjoin",
  		    "username": "Group",
  		    "language_code": "en"
  		  },
  		"chat": {
  		    "id": ` + strconv.Itoa(GroupID) + `,
  		    "title": "Example Group",
  		    "type": "group"
  		  },
  		"date": 1609459260,
  		"new_chat_members": [
  		    {
  		      "id":  ` + strconv.Itoa(int(userid)) + `,
  		      "is_bot": false,
  		      "first_name": "Alice",
  		      "username": "alice_smith",
  		      "language_code": "en"
  		    }
  		  ],
  		"text": "Welcome to Example Group, @alice_smith!",
  		"entities": [
  		    {
  		      "offset": 24,
  		      "length": 11,
  		      "type": "mention"
  		   	 }
  		  		]
  			}
		}
	`

	userid2 := rand.Int63()

	chatmemjoin := `
	{
  "update_id": 123456792,
  "my_chat_member": {
    "chat": {
      "id": ` + strconv.Itoa(GroupID) + `,
      "title": "Test Group",
      "type": "group"
    },
    "from": {
      "id": ` + strconv.Itoa(int(userid2)) + `,
      "is_bot": false,
      "first_name": "toleave",
      "username": "Chatmem"
    },
    "date": 1680300003,
    "new_chat_member": {
      "user": {
        "id": ` + strconv.Itoa(int(userid2)) + `,
        "is_bot": false,
        "first_name": "John",
        "username": "john_doe"
      },
      "status": "member"
    }
  }
}
	`

	//userid2 = 6767490956995604587
	chatmemleave := `
	{
  "update_id": 123456793,
  "my_chat_member": {
    "chat": {
      "id": ` + strconv.Itoa(GroupID) + `,
      "title": "Test Group",
      "type": "group"
    },
    "from": {
      "id": ` + strconv.Itoa(int(userid2)) + `,
      "is_bot": false,
      "first_name": "John",
      "username": "john_doe"
    },
    "date": 1680300004,
    "new_chat_member": {
      "user": {
        "id": ` + strconv.Itoa(int(userid2)) + `,
        "is_bot": false,
        "first_name": "John",
        "username": "john_doe"
      },
      "status": "left"
    }
  }
}

	`
	allupdates := []string{
		commandstart,
		commandhelp,
		commandsendgift,
		newuserjoin,
		chatmemjoin,
		chatmemleave,
		commandcreate,
	}
	singlereq(chatmemjoin, false)
	singlereq(chatmemleave, false)
	for _, update := range allupdates {
		singlereq(update, true)
	}

}

func singlereq(rawjson string, off bool) {
	if off {
		return
	}
	buf := bytes.NewBuffer([]byte(rawjson))
	st := time.Now()
	res, err := http.Post("http://127.0.0.1:5566/", "text/html", buf)
	if err != nil {
		zLogger.Error(err.Error())
	}
	fmt.Println(time.Since(st))
	zLogger.Info(res.Status)
	buf = nil
}

func TestSendup(t *testing.T) {
	sendupdateds()
}

func TestQuotacalc(t *testing.T) {
	var userquotaOld int = 50
	var userquotaNew int = 43
	userconfigs := [][]int{
		[]int{25, 25},
		[]int{30, 20},
		[]int{30, 10, 10},
		[]int{10, 40},
	}

	for _, i := range userconfigs {
		var totalquotaforconfig int = 0
		for _, configquota := range i {

			totalquotaforconfig = totalquotaforconfig + int(float32(userquotaNew)/(float32(userquotaOld)/float32(configquota)))

		}
		fmt.Println(totalquotaforconfig)
	}

}

func TestTiming(t *testing.T) {
	uuidMap := make(map[uuid.UUID]string)
	testuuid := uuid.New()
	syncmap := sync.Map{}
	for i := 0; i < 100000; i++ {
		storeid := uuid.New()
		uuidMap[storeid] = "waefdaewf"
		syncmap.Store(storeid, i)
	}
	syncmap.Store(testuuid, "ss")
	uuidMap[testuuid] = "testuuid"

	// Generate a key to look up

	// Measure lookup time
	start := time.Now()
	ssss := uuidMap[testuuid]
	elapsed := time.Since(start)

	start = time.Now()
	syncmap.Load(testuuid)
	elapsed2 := time.Since(start)

	fmt.Println(ssss)

	fmt.Printf("Lookup time for normal : %v\n", elapsed.Nanoseconds())
	fmt.Printf("Lookup time for syncmap : %v\n", elapsed2.Nanoseconds())
}

func TestBuf(t *testing.T) {
	buf := make(chan string, 1000)

	go func() {
		//newbuf := bytes.NewBuffer([]byte{})

		for {

			val := <-buf
			time.Sleep(2 * time.Microsecond)
			fmt.Println(val)

		}
	}()

	st := time.Now()
	for i := 0; i < 100; i++ {

		buf <- "Hello this is logger test speed"

	}

	fmt.Println(time.Since(st).Nanoseconds(), "for buffering all logs")
	time.Sleep(15 * time.Second)

}
