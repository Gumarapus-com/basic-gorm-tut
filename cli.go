package main

import (
  "context"
  "flag"
  "fmt"
  "time"

  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
  "gorm.io/gorm/logger"
)

const USAGE = `Actions:
 - user_create
   Ex: cli -action=user_create -email=user@example.com -name="User Name"
 - user_select
   Ex: cli -action=user_select -email=user@example.com
 - note_create
   Ex: cli -action=note_create -email=user@example.com -title="Note Title" -text="Note text"
 - note_select
   Ex: cli -action=note_select -email=user@example.com
`

const TIME_FMT = "2006-01-02 15:04:05"

type User struct {
  ID        uint `gorm:"primarykey"`
  CreatedAt time.Time
  Email     string `gorm:"unique"`
  Name      string
}

func (u *User) String() string {
  return fmt.Sprintf(
    "name=%s email=%s created_at=%s",
    u.Name,
    u.Email,
    u.CreatedAt.Format(time.DateTime),
  )
}

type Note struct {
  ID        uint `gorm:"primarykey"`
  CreatedAt time.Time
  Title     string
  Text      string
  // relationship to User
  AuthorID int
  Author   User `gorm:"foreignKey:AuthorID"`
}

func (n *Note) String(author *User) string {
  return fmt.Sprintf("\ttitle=%s\n\ttext=%s\n\tauthor: %s", n.Title, n.Text, author)
}

func main() {
  db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{
    // Disable query log
    Logger: logger.Default.LogMode(logger.Silent),
  })
  if err != nil {
    fmt.Println("Failed to connect to DB")
    return
  }

  // unsure of what context to provide. maybe need to refactor later
  ctx := context.TODO()

  // Create tables
  db.AutoMigrate(&User{}, &Note{})

  //

  action := flag.String("action", "", "valid actions: user_select, user_create, note_select, note_create")

  userEmail := flag.String("email", "", "User's email")
  userName := flag.String("name", "", "User's name")

  noteTitle := flag.String("title", "", "Note title")
  noteText := flag.String("text", "", "Note text")

  flag.Parse()

  // every command need user's email
  if userEmail == nil {
    fmt.Println("Email is required")
    return
  }

  user, err := gorm.G[User](db).Where("email = ?", userEmail).First(ctx)

  switch *action {
  case "user_select":
    if err != nil {
      fmt.Println("Error:", err.Error())
      break
    }
    fmt.Printf("Got user: %s\n", user.String())
  case "user_create":
    if userName == nil {
      fmt.Println("Error: name is required")
      break
    }
    userData := &User{
      CreatedAt: time.Now(),
      Email:     *userEmail,
      Name:      *userName,
    }
    err = gorm.G[User](db).Create(ctx, userData)
    if err != nil {
      fmt.Println("Error:", err.Error())
      break
    }
    fmt.Printf("User created %s\n", userData.String())
  case "note_select":
    if err != nil {
      fmt.Println("Error:", err.Error())
      break
    }

    fmt.Println("User notes:")
    notes, _ := gorm.G[Note](db).Where("author_id = ?", user.ID).Find(ctx)
    for index, note := range notes {
      fmt.Printf("%2d %s\n", index+1, note.String(&user))
    }
  case "note_create":
    if err != nil {
      fmt.Println("Error:", err.Error())
      break
    }

    if noteTitle == nil {
      fmt.Println("Error: note title is required")
      break
    }
    if noteText == nil {
      fmt.Println("Error: note text is required")
      break
    }
    _ = gorm.G[Note](db).Create(ctx, &Note{
      CreatedAt: time.Now(),
      Author:    user,
      Title:     *noteTitle,
      Text:      *noteText,
    })
    fmt.Printf("Note created. title=%s\n", *noteTitle)
  default:
    fmt.Printf("Error: Invalid action \"%s\". ", *action)
    fmt.Println(USAGE)
  }
}
