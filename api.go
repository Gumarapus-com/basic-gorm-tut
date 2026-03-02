package main

import (
  "context"
  "fmt"
  "log"
  "time"

  "github.com/gofiber/fiber/v2"
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
  "gorm.io/gorm/clause"
)

var db *gorm.DB
var err error

type User struct {
  ID        uint `gorm:"primaryKey"`
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
  ID        uint `gorm:"primaryKey"`
  CreatedAt time.Time
  Title     string
  Text      string
  // relationship to User
  AuthorID int  `json:"author_id"`
  Author   User `gorm:"foreignKey:AuthorID"`
}

func (n *Note) String(author *User) string {
  return fmt.Sprintf("\ttitle=%s\n\ttext=%s\n\tauthor: %s", n.Title, n.Text, author)
}

func main() {
  // Initialize the database
  db, err = gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
  if err != nil {
    panic("failed to connect database")
  }

  // Migrate the schema
  db.AutoMigrate(&User{}, &Note{})

  // Initialize Fiber app
  app := fiber.New()

  // Define routes
  app.Post("/users", CreateUser)
  app.Get("/users/:email", GetUser)
  app.Post("/notes/", CreateNote)
  app.Get("/notes/:email", GetUserNote)

  // Start the server
  log.Fatal(app.Listen(":3000"))
}

func CreateUser(c *fiber.Ctx) error {
  var user User

  // Parse request json payload
  if err := c.BodyParser(&user); err != nil {
    // Bad request
    return c.Status(400).JSON(fiber.Map{
      "error": err.Error(),
    })
  }

  if err := gorm.G[User](db).Create(c.Context(), &user); err != nil {
    return c.Status(500).JSON(fiber.Map{
      "error": err.Error(),
    })
  }
  return c.JSON(user)
}

func GetUser(c *fiber.Ctx) error {
  email := c.Params("email")
  user, err := gorm.G[User](db).Where("email = ?", email).First(c.Context())

  if err != nil {
    return c.Status(404).JSON(fiber.Map{"error": err.Error()})
  }
  return c.JSON(user)
}

func CreateNote(c *fiber.Ctx) error {
  var note Note

  if err := c.BodyParser(&note); err != nil {
    return c.Status(400).JSON(fiber.Map{
      "error": "Cannot parse JSON",
    })
  }

  user, err := gorm.G[User](db).Where("id = ?", note.AuthorID).First(c.Context())
  if err != nil {
    return c.Status(400).JSON(fiber.Map{
      "error":     err.Error(),
      "message":   "Get user error",
      "author_id": note.AuthorID,
    })
  }

  if err := gorm.G[Note](db).Create(context.Background(), &note); err != nil {
    // 500 - internal server error
    return c.Status(500).JSON(fiber.Map{
      "error": err.Error(),
    })
  }
  note.Author = user

  return c.JSON(note)
}

func GetUserNote(c *fiber.Ctx) error {
  email := c.Params("email")

  user, err := gorm.G[User](db).Where("email = ?", email).First(c.Context())
  if err != nil {
    return c.Status(404).JSON(fiber.Map{"error": err.Error()})
  }

  notes, err := gorm.G[Note](db).Joins(
    clause.JoinTarget{Association: "Author"}, nil,
  ).Where("author_id = ?", user.ID).Find(c.Context())

  if err != nil {
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
  }
  return c.JSON(notes)
}
