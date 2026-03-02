package main

import (
  "context"
  "fmt"
  "time"

  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
)

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
  db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
  if err != nil {
    panic("Failed to connect to DB")
  }

  // unsure of what context to provide. maybe need to refactor later
  ctx := context.TODO()

  // Create tables
  db.AutoMigrate(&User{}, &Note{})

  // Insert user
  userData := &User{
    CreatedAt: time.Now(),
    Email:     "myemail@example.com",
    Name:      "First user",
  }
  err = gorm.G[User](db).Create(ctx, userData)
  if err != nil {
    panic("Fail to create user")
  }
  fmt.Printf("Create user %s\n", userData.String())

  // Select user
  user, err := gorm.G[User](db).Where("email = ?", userData.Email).First(ctx)
  if err != nil {
    panic("User not found")
  }
  fmt.Printf("\nSelect user %s\n", user.String())

  // Insert a note
  _ = gorm.G[Note](db).Create(ctx, &Note{
    CreatedAt: time.Now(),
    Author:    user,
    Title:     "First note",
    Text:      "First note text",
  })

  // Insert more note
  _ = gorm.G[Note](db).Create(ctx, &Note{
    CreatedAt: time.Now(),
    Author:    user,
    Title:     "Second note",
    Text:      "Second note text",
  })

  // Select user
  notes, _ := gorm.G[Note](db).Where("author_id = ?", user.ID).Find(ctx)

  fmt.Println("\nShowing user notes")
  for num, note := range notes {
    fmt.Printf("%2d. %s\n", num+1, note.String(&user))
  }
}
