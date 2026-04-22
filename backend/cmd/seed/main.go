package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type player struct {
	name         string
	dateOfBirth  *time.Time
	role         *string
	battingStyle *string
	bowlingStyle *string
	country      string
	playerImg    string
}

func sp(s string) *string { return &s }

func dob(s string) *time.Time {
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return nil
	}
	return &t
}

var players = []player{
	{name: "Vicky Ostwal", dateOfBirth: dob("2002-09-01T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "India", playerImg: "https://h.cricapi.com/img/players/b9701b53-3b9e-4eda-93b1-3543809674d4.jpg"},
	{name: "Romario Shepherd", dateOfBirth: dob("1994-11-26T00:00:00"), role: sp("Bowling Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "West Indies", playerImg: "https://h.cricapi.com/img/players/7fd4fa20-bc49-4337-ae8a-540b67cb011d.jpg"},
	{name: "Jacob Bethell", dateOfBirth: dob("2003-10-23T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "England", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Rajat Patidar", dateOfBirth: dob("1993-06-01T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Tim David", dateOfBirth: dob("1996-03-16T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "Australia", playerImg: "https://h.cricapi.com/img/players/69718248-2a3a-45d4-9a2a-6f4dfc6942a5.jpg"},
	{name: "Jordan Cox", dateOfBirth: dob("2000-10-21T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: nil, country: "England", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Mangesh Yadav", dateOfBirth: dob("2002-10-10T00:00:00"), role: sp("Bowling Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Left-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Satvik Deswal", dateOfBirth: nil, role: sp("Bowling Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Josh Hazlewood", dateOfBirth: dob("1991-01-08T00:00:00"), role: sp("Bowler"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "Australia", playerImg: "https://h.cricapi.com/img/players/2190c28d-1712-4fd2-ae44-9ac54319fc21.jpg"},
	{name: "Venkatesh Iyer", dateOfBirth: dob("1994-12-25T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/c3236037-6694-404a-8f01-9ad880564ea9.jpg"},
	{name: "Abhinandan Singh", dateOfBirth: dob("1997-03-30T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Swapnil Singh", dateOfBirth: dob("1991-01-22T00:00:00"), role: sp("Bowling Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Vihaan Malhotra", dateOfBirth: dob("2007-01-01T00:00:00"), role: sp("Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Jacob Duffy", dateOfBirth: dob("1994-08-02T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "New Zealand", playerImg: "https://h.cricapi.com/img/players/6e05cbea-6a4b-42ef-8e2e-b44e4b434c2f.jpg"},
	{name: "Yash Dayal", dateOfBirth: dob("1997-12-13T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Bhuvneshwar Kumar", dateOfBirth: dob("1990-02-05T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/3f3ecf51-8411-4046-9477-18c0fe3da6ac.jpg"},
	{name: "Virat Kohli", dateOfBirth: dob("1988-11-05T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/c61d247d-7f77-452c-b495-2813a9cd0ac4.jpg"},
	{name: "Rasikh Salam Dar", dateOfBirth: dob("2000-04-05T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Jitesh Sharma", dateOfBirth: dob("1993-10-22T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/players/e28de73b-f5df-49eb-bdf6-c50471319404.jpg"},
	{name: "Devdutt Padikkal", dateOfBirth: dob("2000-07-07T00:00:00"), role: sp("Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/players/74c6584a-45a5-4781-a5e7-c0c9340da954.jpg"},
	{name: "Krunal Pandya", dateOfBirth: dob("1991-03-24T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "India", playerImg: "https://h.cricapi.com/img/players/81c09c1b-1b1d-4e87-9eaa-d7d0a89a6159.jpg"},
	{name: "Philip Salt", dateOfBirth: dob("1996-08-28T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "England", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Kanishk Chouhan", dateOfBirth: dob("2006-09-26T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Nuwan Thushara", dateOfBirth: dob("1994-08-06T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "Sri Lanka", playerImg: "https://h.cricapi.com/img/players/eb911b3b-10b3-40f8-9294-e63ba030af83.jpg"},
	{name: "Suyash Sharma", dateOfBirth: dob("2003-05-15T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm legbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Praful Hinge", dateOfBirth: dob("2002-01-18T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Eshan Malinga", dateOfBirth: dob("2001-02-04T00:00:00"), role: sp("Bowler"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "Sri Lanka", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Heinrich Klaasen", dateOfBirth: dob("1991-07-30T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "South Africa", playerImg: "https://h.cricapi.com/img/players/7d662361-8f12-4a52-9a4b-01e8734bd31d.jpg"},
	{name: "Nitish Kumar Reddy", dateOfBirth: dob("2003-05-26T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Harshal Patel", dateOfBirth: dob("1990-11-23T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/15bccb4f-8c3c-4da6-acd9-551782278bfd.jpg"},
	{name: "Sakib Hussain", dateOfBirth: dob("2004-12-14T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Jack Edwards", dateOfBirth: dob("2000-04-19T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "Australia", playerImg: "https://h.cricapi.com/img/players/6d751be3-3023-43b9-8415-8259d19689fd.jpg"},
	{name: "Krains Fuletra", dateOfBirth: nil, role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Kamindu Mendis", dateOfBirth: dob("1998-09-30T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "Sri Lanka", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Shivang Kumar", dateOfBirth: dob("2002-05-26T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "David Payne", dateOfBirth: dob("1991-02-15T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm fast-medium"), country: "England", playerImg: "https://h.cricapi.com/img/players/7cb23ef6-2cd5-4aeb-9047-a5054d220f98.jpg"},
	{name: "Onkar Tarmale", dateOfBirth: nil, role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Aniket Verma", dateOfBirth: dob("2002-02-05T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Smaran Ravichandran", dateOfBirth: dob("2003-05-05T00:00:00"), role: sp("Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Shivam Mavi", dateOfBirth: dob("1998-11-26T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/f00ccdba-6137-4852-a490-ba782044b866.jpg"},
	{name: "Abhishek Sharma", dateOfBirth: dob("2000-09-04T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "India", playerImg: "https://h.cricapi.com/img/players/83fe6ab6-d63c-420e-9416-a93d59a9a964.jpg"},
	{name: "Amit Kumar", dateOfBirth: nil, role: nil, battingStyle: nil, bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Pat Cummins", dateOfBirth: dob("1993-05-08T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast"), country: "Australia", playerImg: "https://h.cricapi.com/img/players/13b3d56e-0fba-4d31-a174-d211211404e2.jpg"},
	{name: "Travis Head", dateOfBirth: dob("1993-12-29T00:00:00"), role: sp("Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "Australia", playerImg: "https://h.cricapi.com/img/players/d6720897-e81d-4703-bf33-d2b60c973d30.jpg"},
	{name: "Harsh Dubey", dateOfBirth: dob("2002-07-23T00:00:00"), role: sp("Bowling Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Salil Arora", dateOfBirth: dob("2002-11-07T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Brydon Carse", dateOfBirth: dob("1995-07-31T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast"), country: "England", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Ishan Kishan", dateOfBirth: dob("1998-07-18T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/players/9ae68636-48d9-488f-8adc-e62a19815d85.jpg"},
	{name: "Jaydev Unadkat", dateOfBirth: dob("1991-10-18T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/f589d2a9-4e44-464d-9741-e224354a1370.jpg"},
	{name: "Liam Livingstone", dateOfBirth: dob("1993-08-04T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm legbreak"), country: "England", playerImg: "https://h.cricapi.com/img/players/d1c0a448-6ee6-473a-90a2-e6a486698a8a.jpg"},
	{name: "Zeeshan Ansari", dateOfBirth: dob("1999-12-16T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm legbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Finn Allen", dateOfBirth: dob("1999-04-22T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: nil, country: "New Zealand", playerImg: "https://h.cricapi.com/img/players/08924d57-b4d3-4ed0-903d-087a8384be12.jpg"},
	{name: "Varun Chakaravarthy", dateOfBirth: dob("1991-08-29T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm legbreak"), country: "India", playerImg: "https://h.cricapi.com/img/players/702254aa-2764-4fe4-b28e-20336a0ab069.jpg"},
	{name: "Blessing Muzarabani", dateOfBirth: dob("1996-10-02T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "Zimbabwe", playerImg: "https://h.cricapi.com/img/players/50e38c9d-bf39-44b8-b5e6-0c9f36b8cbdf.jpg"},
	{name: "Manish Pandey", dateOfBirth: dob("1989-09-10T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/306f6b88-0442-476a-be55-1500b3441051.jpg"},
	{name: "Umran Malik", dateOfBirth: dob("1999-11-22T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast"), country: "India", playerImg: "https://h.cricapi.com/img/players/b367db88-a73e-445f-8cdd-3f6335617b07.jpg"},
	{name: "Vaibhav Arora", dateOfBirth: dob("1997-12-14T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Harshit Rana", dateOfBirth: dob("2001-12-22T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Tejasvi Dahiya", dateOfBirth: dob("2002-04-18T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Angkrish Raghuvanshi", dateOfBirth: dob("2004-06-05T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Ajinkya Rahane", dateOfBirth: dob("1988-06-06T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/cdc2646c-1fe7-4bc2-b26e-42453f45a212.jpg"},
	{name: "Cameron Green", dateOfBirth: dob("1999-06-03T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "Australia", playerImg: "https://h.cricapi.com/img/players/abe58090-e933-4888-8538-64f4611ee259.jpg"},
	{name: "Sunil Narine", dateOfBirth: dob("1988-05-26T00:00:00"), role: sp("Bowling Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "West Indies", playerImg: "https://h.cricapi.com/img/players/ef2e9c01-41f2-4349-a945-45d8b11dbd19.jpg"},
	{name: "Rinku Singh", dateOfBirth: dob("1997-10-12T00:00:00"), role: sp("Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/players/11864f31-b766-4766-9cbc-5026015ace48.jpg"},
	{name: "Akash Deep", dateOfBirth: dob("1996-12-15T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Kartik Tyagi", dateOfBirth: dob("2000-11-08T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast"), country: "India", playerImg: "https://h.cricapi.com/img/players/ff17824f-0e1c-4eb7-acd3-8b5d116934bc.jpg"},
	{name: "Rovman Powell", dateOfBirth: dob("1993-07-23T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "West Indies", playerImg: "https://h.cricapi.com/img/players/c699ec2b-a5f6-4ee0-a42d-79391b65b293.jpg"},
	{name: "Sarthak Ranjan", dateOfBirth: dob("1996-09-25T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Rahul Tripathi", dateOfBirth: dob("1991-03-02T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/fb313a9a-0a14-499d-a8d1-7561bccea8cb.jpg"},
	{name: "Daksh Kamra", dateOfBirth: nil, role: nil, battingStyle: nil, bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Matheesha Pathirana", dateOfBirth: dob("2002-12-18T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast"), country: "Sri Lanka", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Prashant Solanki", dateOfBirth: dob("2000-02-22T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm legbreak"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Rachin Ravindra", dateOfBirth: dob("1999-11-18T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "New Zealand", playerImg: "https://h.cricapi.com/img/players/e6d28951-8524-43ca-95cf-9e73ae68bc12.jpg"},
	{name: "Tim Seifert", dateOfBirth: dob("1994-12-14T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: nil, country: "New Zealand", playerImg: "https://h.cricapi.com/img/players/812fc568-a932-4f29-a1c4-b7c929e6a248.jpg"},
	{name: "Navdeep Saini", dateOfBirth: dob("1992-11-23T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast"), country: "India", playerImg: "https://h.cricapi.com/img/players/80193c8f-687d-47c3-a7e9-b098a83c7812.jpg"},
	{name: "Anukul Roy", dateOfBirth: dob("1998-11-30T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Left-arm orthodox"), country: "India", playerImg: "https://h.cricapi.com/img/players/2958faf6-f7d1-4150-aa9e-e261a629ba0d.jpg"},
	{name: "Saurabh Dubey", dateOfBirth: dob("1998-01-23T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Ramandeep Singh", dateOfBirth: dob("1997-04-13T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm medium"), country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Raghu Sharma", dateOfBirth: dob("1993-03-11T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Will Jacks", dateOfBirth: dob("1998-11-21T00:00:00"), role: sp("Batting Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "England", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Mayank Markande", dateOfBirth: dob("1997-11-11T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm legbreak"), country: "India", playerImg: "https://h.cricapi.com/img/players/c06c697c-9575-4bb0-b1f0-27a56a2f3898.jpg"},
	{name: "Tilak Varma", dateOfBirth: dob("2002-11-08T00:00:00"), role: sp("Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/players/aefd9074-da5a-408d-94f1-3be0e79d3b82.jpg"},
	{name: "Mohammed Salahuddin Izhar", dateOfBirth: nil, role: nil, battingStyle: nil, bowlingStyle: nil, country: "India", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Ryan Rickelton", dateOfBirth: dob("1996-07-11T00:00:00"), role: sp("WK-Batsman"), battingStyle: sp("Left Handed Bat"), bowlingStyle: nil, country: "South Africa", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "AM Ghazanfar", dateOfBirth: dob("2006-03-20T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "Afghanistan", playerImg: "https://h.cricapi.com/img/icon512.png"},
	{name: "Trent Boult", dateOfBirth: dob("1989-07-22T00:00:00"), role: sp("Bowler"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Left-arm fast-medium"), country: "New Zealand", playerImg: "https://h.cricapi.com/img/players/6004fd3f-2264-470d-b39f-340d530b19b3.jpg"},
	{name: "Rohit Sharma", dateOfBirth: dob("1987-04-30T00:00:00"), role: sp("Batsman"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm offbreak"), country: "India", playerImg: "https://h.cricapi.com/img/players/03bda674-3916-4d64-952e-00a6c19c01e1.jpg"},
	{name: "Shardul Thakur", dateOfBirth: dob("1991-10-16T00:00:00"), role: sp("Bowling Allrounder"), battingStyle: sp("Right Handed Bat"), bowlingStyle: sp("Right-arm fast-medium"), country: "India", playerImg: "https://h.cricapi.com/img/players/c1257351-51ca-446f-b7f5-01fe92e75076.jpg"},
}

func symbolFromName(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, " ", "_"))
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	inserted := 0
	skipped := 0
	for _, p := range players {
		id := uuid.New()
		symbol := symbolFromName(p.name)
		tag, err := pool.Exec(ctx, `
			INSERT INTO assets (id, symbol, name, description, nationality, role, date_of_birth, batting_style, bowling_style, player_img, supply_used)
			VALUES ($1, $2, $3, '', $4, $5, $6, $7, $8, $9, 0)
			ON CONFLICT (symbol) DO NOTHING
		`, id, symbol, p.name, p.country, p.role, p.dateOfBirth, p.battingStyle, p.bowlingStyle, p.playerImg)
		if err != nil {
			log.Printf("insert %s: %v", p.name, err)
			continue
		}
		if tag.RowsAffected() == 0 {
			skipped++
		} else {
			inserted++
		}
	}
	fmt.Printf("done: %d inserted, %d skipped (conflict)\n", inserted, skipped)
}
