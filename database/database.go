package database

// package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type User struct {
	UserName string `dynamodbav:"UserName"`
	Name     string `dynamodbav:"Name"`
	IsAdmin  bool   `dynamodbav:"IsAdmin"`
	Status   string `dynamodbav:"Status"`
}

type Book struct {
	BookID string `dynamodbav:"BookID"`
	Title  string `dynamodbav:"Title"`
	Author string `dynamodbav:"Author"`
	Active bool   `dynamodbav:"Active"`
}

type BookType string

const (
	AudioBook   BookType = "audio"
	RegularBook BookType = "regular"
)

type ReadingProgress struct {
	UserName   string   `dynamodbav:"UserName"`
	BookID     string   `dynamodbav:"BookID"`
	Progress   int      `dynamodbav:"Progress"`
	Type       BookType `dynamodbav:"Type"`
	TotalPages int      `dynamodbav:"TotalPages"`
	PageNumber int      `dynamodbav:"PageNumber"`
}

func AWSsession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		log.Fatalf("Failed to create session: %s", err)
	}
	return sess
}

func IsUserExists(username string) bool {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	input := &dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(username),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	if result.Item == nil {
		return false
	}

	var user User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		log.Fatalf("Failed to unmarshal DynamoDB item: %s", err)
	}

	return true
}

func CreateUser(userName, name string, isAdmin bool) User {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	exists := IsUserExists(userName)
	if exists {
		fmt.Printf("User %s already exists\n", userName)
		return User{UserName: userName, Name: name, IsAdmin: isAdmin}
	}

	item, err := dynamodbattribute.MarshalMap(User{
		UserName: userName,
		Name:     name,
		IsAdmin:  isAdmin,
	})
	if err != nil {
		log.Fatalf("Failed to marshal User: %s", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Users"),
		Item:      item,
	}

	// Put the item into the DynamoDB table
	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Failed to put User item into DynamoDB table: %s", err)
	}

	fmt.Printf("Successfully added user: %s\n", userName)
	return User{UserName: userName, Name: name, IsAdmin: isAdmin}
}

func IsUserAdmin(username string) bool {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	input := &dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(username),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	if result.Item == nil {
		return false
	}

	var user User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		log.Fatalf("Failed to unmarshal DynamoDB item: %s", err)
	}

	return user.IsAdmin
}

func UserList() []User {
	// Create a new AWS session
	sess := AWSsession()

	// Create a DynamoDB client from just created session
	svc := dynamodb.New(sess)

	// Define the query input
	input := &dynamodb.ScanInput{
		TableName: aws.String("Users"), // Replace with your DynamoDB table name
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	var userList []User

	// Unmarshal the Items field in the result value to the User struct
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &userList)
	if err != nil {
		log.Fatalf("Failed to unmarshal Query result items, %v", err)
	}

	return userList
}

func IsUserBelongsToClub(username string) bool {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	input := &dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(username),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	if result.Item == nil {
		return false
	}

	var user User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		log.Fatalf("Failed to unmarshal DynamoDB item: %s", err)
	}

	return true
}

func AddBook(title string, author string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	// Step 1: Make all existing books inactive
	// Scan the table to find all books that are currently active
	filt := expression.Name("Active").Equal(expression.Value(true))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		log.Fatalf("Got error building expression: %s", err)
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String("Books"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	result, err := svc.Scan(scanInput)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	// Update each active book to be inactive
	for _, item := range result.Items {
		var book Book
		err = dynamodbattribute.UnmarshalMap(item, &book)
		if err != nil {
			log.Fatalf("Failed to unmarshal Book, %v", err)
		}

		// Update the book to inactive
		updateInput := &dynamodb.UpdateItemInput{
			TableName: aws.String("Books"),
			Key: map[string]*dynamodb.AttributeValue{
				"BookID": {
					S: aws.String(book.BookID),
				},
			},
			UpdateExpression: aws.String("set Active = :val"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":val": {
					BOOL: aws.Bool(false),
				},
			},
		}

		_, err = svc.UpdateItem(updateInput)
		if err != nil {
			log.Fatalf("Got error calling UpdateItem: %s", err)
		}
	}

	// Step 2: Add the new book as active
	// Use the current timestamp as BookID
	bookID := strconv.FormatInt(time.Now().UnixNano(), 10)

	newBook := Book{
		BookID: bookID,
		Title:  title,
		Author: author,
		Active: true,
	}

	newBookItem, err := dynamodbattribute.MarshalMap(newBook)
	if err != nil {
		log.Fatalf("Got error marshalling new book item: %s", err)
	}

	putInput := &dynamodb.PutItemInput{
		TableName: aws.String("Books"),
		Item:      newBookItem,
	}

	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
	}

	fmt.Println("Successfully added book and updated existing books status")
}

func GetCurrentBook() Book {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	filt := expression.Name("Active").Equal(expression.Value(true))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		log.Fatalf("Got error building expression: %s", err)
	}

	// Prepare the ScanInput
	input := &dynamodb.ScanInput{
		TableName:                 aws.String("Books"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		// Limit:                     aws.Int64(1),
	}

	result, err := svc.Scan(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	if len(result.Items) == 0 {
		log.Println("No active book found")
		return Book{}
	}

	var book Book
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &book)
	if err != nil {
		log.Fatalf("Failed to unmarshal DynamoDB item to Book: %s", err)
	}

	return book
}

func SetProgress(progress ReadingProgress) {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	if progress.UserName == "" || progress.BookID == "" {
		log.Println("Invalid reading progress data: missing key attributes.")
		return
	}

	av, err := dynamodbattribute.MarshalMap(progress)
	if err != nil {
		log.Fatalf("Failed to marshal reading progress for DynamoDB: %s", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("ReadingProgress"),
		Item:      av,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Failed to put reading progress item into DynamoDB: %s", err)
	}

	log.Printf("Updated reading progress for user '%s' on book '%s'.\n", progress.UserName, progress.BookID)
}

func GroupProgress() string {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	activeBook := GetCurrentBook()
	if activeBook.BookID == "" {
		return "No active book found."
	}

	keyCond := expression.Key("BookID").Equal(expression.Value(activeBook.BookID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		log.Fatalf("Failed to build expression: %s", err)
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String("ReadingProgress"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	}

	result, err := svc.Query(queryInput)
	if err != nil {
		log.Fatalf("Failed to query reading progress: %s", err)
	}

	var progresses []ReadingProgress
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &progresses)
	if err != nil {
		log.Fatalf("Failed to unmarshal reading progress results: %s", err)
	}

	sort.Slice(progresses, func(i, j int) bool {
		return progresses[i].Progress > progresses[j].Progress
	})

	var groupProgress string
	for _, progress := range progresses {
		user := GetUserDetails(progress.UserName)
		groupProgress += fmt.Sprintf("%s: %d%%\n", user.Name, progress.Progress)
	}

	if groupProgress == "" {
		return "No users have set their progress yet."
	}

	return groupProgress
}

func GetUserDetails(userName string) User {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	input := &dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(userName),
			},
		},
	}

	// Execute the GetItem operation
	result, err := svc.GetItem(input)
	if err != nil {
		log.Fatalf("Failed to get user details for UserName '%s': %s", userName, err)
	}

	if result.Item == nil {
		log.Printf("User '%s' not found.", userName)
		return User{} // Return an empty User struct if not found
	}

	// Unmarshal the result item into a User struct
	var user User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		log.Fatalf("Failed to unmarshal user details: %s", err)
	}

	return user
}

func RemoveBook(bookID string) {
	sess := AWSsession() // Assuming AWSsession() is your function to create and return an AWS session
	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("Books"), // Replace with your actual table name
		Key: map[string]*dynamodb.AttributeValue{
			"BookID": {
				S: aws.String(bookID),
			},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		log.Fatalf("Failed to delete book with ID '%s': %s", bookID, err)
	} else {
		log.Printf("Book with ID '%s' removed successfully.", bookID)
	}
}

func AddUser(userName string, name string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	exists := IsUserExists(userName)
	if exists {
		fmt.Printf("User %s already exists\n", userName)
		return
	}

	item, err := dynamodbattribute.MarshalMap(User{
		UserName: userName,
		Name:     name,
		IsAdmin:  false,
	})
	if err != nil {
		log.Fatalf("Failed to marshal User: %s", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Users"),
		Item:      item,
	}

	// Put the item into the DynamoDB table
	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Failed to put User item into DynamoDB table: %s", err)
	}

	fmt.Printf("Successfully added user: %s\n", userName)
}

func RemoveUser(userName string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("Users"), // Replace with your actual table name
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(userName),
			},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		log.Fatalf("Failed to delete user with UserName '%s': %s", userName, err)
	} else {
		log.Printf("User with UserName '%s' removed successfully.", userName)
	}
}

func BookList() []Book {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	input := &dynamodb.ScanInput{
		TableName: aws.String("Books"),
	}

	result, err := svc.Scan(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	var bookList []Book

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &bookList)
	if err != nil {
		log.Fatalf("Failed to unmarshal Query result items, %v", err)
	}

	return bookList
}

func UserStatus(userName string) string {
	sess := AWSsession()
	svc := dynamodb.New(sess)
	key := map[string]*dynamodb.AttributeValue{
		"UserName": {
			S: aws.String(userName),
		},
	}

	input := &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String("Users"),
	}

	result, _ := svc.GetItem(input)
	var user User
	dynamodbattribute.UnmarshalMap(result.Item, &user)
	return user.Status
}

func SetUserStatus(userName string, status string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	key := map[string]*dynamodb.AttributeValue{
		"UserName": {
			S: aws.String(userName),
		},
	}

	updateExpression := "SET #st = :s"

	expressionAttributeNames := map[string]*string{
		"#st": aws.String("Status"),
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":s": {
			S: aws.String(status),
		},
	}
	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String("Users"),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              aws.String("UPDATED_NEW"),
	}

	result, err := svc.UpdateItem(input)
	fmt.Println("result!", result)
	if err != nil {
		fmt.Printf("Failed to update item: %v\n", err)
		return
	}

	fmt.Println("User status updated successfully")
}

func UserProgress(userName string) *ReadingProgress {
	sess := session.Must(session.NewSession()) // Assume AWSsession() initializes a session
	svc := dynamodb.New(sess)

	activeBook := GetCurrentBook() // Assume GetCurrentBook() returns the current book details
	if activeBook.BookID == "" {
		log.Fatal("No active book found.")
	}

	keyCond := expression.Key("BookID").Equal(expression.Value(activeBook.BookID)).
		And(expression.Key("UserName").Equal(expression.Value(userName)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		log.Fatalf("Failed to build expression: %s", err)
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String("ReadingProgress"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	}

	result, err := svc.Query(queryInput)
	if err != nil {
		log.Fatalf("Failed to query reading progress: %s", err)
	}

	var progresses []ReadingProgress
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &progresses)
	if err != nil {
		log.Fatalf("Failed to unmarshal reading progress results: %s", err)
	}

	if len(progresses) == 0 {
		return nil // No progress found, return nil
	}

	// Assuming there is only one record per user per book
	return &progresses[0]
}

// func main() {
// 	readingProgress := ReadingProgress{
// 		UserName: "ads",
// 		BookID:   "123",
// 		Type:     "asd",
// 	}
// 	fmt.Println(readingProgress)
// }
