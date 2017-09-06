package main

import (
	"fmt"
	"database/sql"
	"strings"
	"bytes"
	"strconv"


	"github.com/nlopes/slack"
	_ "github.com/lib/pq"
)

import private "./secrets"




func make_query(query string, t string) (string){
	db_table := make(map[string]string)
	db_table["wind speed"] = "max_sus_wind"
	db_table["pressure"] = "min_cent_pressure"
	db_table["speed"] = "speed"

	s := strings.Split(query, "AA")
	fmt.Println(t)
    list := []string{s[0], db_table[t], s[1]}
    var str bytes.Buffer

    for _, l := range list {
        str.WriteString(l)
    }
	fmt.Println(str.String())
	fmt.Println(db_table[t])
    return str.String()
 }

func gather_info(year int, column string)(string){

 psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    private.Host, private.Port, private.User, private.Password, private.Dbname)
  db, err := sql.Open("postgres", psqlInfo)
  if err != nil {
	fmt.Println("i found the fucking error")
    panic(err)
  }
  defer db.Close()

  err = db.Ping()
  if err != nil {
	fmt.Println("error in db.ping")
    panic(err)
  }

  fmt.Println("Successfully connected!")
  switch {
  case strings.Contains(column, "lowest"):
	text2 := "pressure" 
 	joined_query := make_query("select current_name, min_pressure, max_sus_wind from (select distinct on (storm.id)storm.id,  v.AA as min_pressure, v.current_name, v.max_sus_wind from tracker_storm storm inner join tracker_advisory v on storm.id = v.stormid_id where storm.year >= 2017 order by storm.id, v.min_cent_pressure) t order by min_pressure limit 10", text2)
	rows, err := db.Query(joined_query)
		if err != nil {
		fmt.Println(err)
		}
		defer rows.Close()
		 fmt.Println("starting to iterate")
		 
		
			if err != nil {
		fmt.Print(err)
		return "there was an error in your query"
	}
		response := ""
		 		 for rows.Next() {
					var min_pressure int
					var max_wind int
					
					var current_name string
					err := rows.Scan(&current_name, &min_pressure, &max_wind)
					
						if err != nil {
					fmt.Print(err)
					return "there was an error in your query"

				}
					press := strconv.Itoa(min_pressure)
					wind := strconv.Itoa(max_wind)
					fmt.Printf("%v", current_name)
					fmt.Println(year)
					response +=  current_name+" "+press+" "+wind +"\n"
					
			}
				return response

		




  case strings.Contains(column, "active storms"):
	rows, err := db.Query("SELECT stormid FROM tracker_storm WHERE active = true")
			if err != nil {
		fmt.Println(err)
		}
		defer rows.Close()
		 fmt.Println("starting to iterate")
		  		response := ""
		 		 for rows.Next() {
		
					var current_name string
					err := rows.Scan(&current_name)
						if err != nil {
					fmt.Print(err)
					return "there was an error in your query"
				}
					
					fmt.Printf("%v", current_name)
					fmt.Println(year)
					response +=  "\n"+ current_name
					
			}
				return response
	
	case strings.Contains(column, "storm data"):
		text2 := strings.Split(column,"storm data ")[1]
		return text2
	
					
  } 



	return "nothing returned"
}	



func main() {

	token := "xoxb-232111251922-pBpm1eT81fqjyGJVKwD1LsCI"
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Print("Event Received: ")
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				fmt.Println("Connection counter:", ev.ConnectionCount)

			case *slack.MessageEvent:
				fmt.Printf("Message: %v\n", ev)
				info := rtm.GetInfo()
				prefix := fmt.Sprintf("<@%s> ", info.User.ID)
				fmt.Printf(prefix)

				if ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) { //checks to see if the event was directed toward bot
					text:= strings.Split(ev.Text, "> " )[1] 
					fmt.Println(text) //prints message text
					reply := gather_info(2017,text)
					rtm.SendMessage(rtm.NewOutgoingMessage(reply, ev.Channel))

				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				//Take no action
			}
		}
	}
}