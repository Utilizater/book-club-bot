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
	FullName string `dynamodbav:"FullName"`
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

func tableName(table string) string {
	environment := os.Getenv("ENV")

	if environment == "" {
		log.Fatal("There is no environment")
	}

	type TableConfig = map[string]string

	var tablesPerEnv = map[string]TableConfig{
		"users": {
			"prod": "Users",
			"dev":  "Users_dev",
		},
		"books": {
			"prod": "Books",
			"dev":  "Books_dev",
		},
		"reading_progress": {
			"prod": "ReadingProgress",
			"dev":  "ReadingProgress_dev",
		},
	}

	return tablesPerEnv[table][environment]
}

func IsUserExists(username string) bool {
	sess := AWSsession()
	svc := dynamodb.New(sess)
	usersTable := tableName("users")
	input := &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
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
		return User{UserName: userName, FullName: name, IsAdmin: isAdmin}
	}

	item, err := dynamodbattribute.MarshalMap(User{
		UserName: userName,
		FullName: name,
		IsAdmin:  isAdmin,
	})
	if err != nil {
		log.Fatalf("Failed to marshal User: %s", err)
	}
	usersTable := tableName("users")
	input := &dynamodb.PutItemInput{
		TableName: aws.String(usersTable),
		Item:      item,
	}

	// Put the item into the DynamoDB table
	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Failed to put User item into DynamoDB table: %s", err)
	}

	fmt.Printf("Successfully added user: %s\n", userName)
	return User{UserName: userName, FullName: name, IsAdmin: isAdmin}
}

func IsUserAdmin(username string) bool {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	usersTable := tableName("users")
	input := &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
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
	sess := AWSsession()
	svc := dynamodb.New(sess)

	usersTable := tableName("users")
	input := &dynamodb.ScanInput{
		TableName: aws.String(usersTable),
	}
	result, err := svc.Scan(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
	}

	var userList []User

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &userList)
	if err != nil {
		log.Fatalf("Failed to unmarshal Query result items, %v", err)
	}

	return userList
}

func IsUserBelongsToClub(username string) bool {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	usersTable := tableName("users")
	input := &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
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

func AddBook(title string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	// Step 1: Make all existing books inactive
	// Scan the table to find all books that are currently active
	filt := expression.Name("Active").Equal(expression.Value(true))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		log.Fatalf("Got error building expression: %s", err)
	}

	booksTable := tableName("books")
	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(booksTable),
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

		booksTable := tableName("books")
		updateInput := &dynamodb.UpdateItemInput{
			TableName: aws.String(booksTable),
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
		Active: true,
	}

	newBookItem, err := dynamodbattribute.MarshalMap(newBook)
	if err != nil {
		log.Fatalf("Got error marshalling new book item: %s", err)
	}

	booksTable = tableName("books")
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(booksTable),
		Item:      newBookItem,
	}

	_, err = svc.PutItem(putInput)
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
	}

	fmt.Println("Successfully added book and updated existing books status")
}

func UpdateBookAuthor(bookID, author string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)
	booksTable := tableName("books")

	// Prepare the update expression
	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(booksTable),
		Key: map[string]*dynamodb.AttributeValue{
			"BookID": {
				S: aws.String(bookID),
			},
		},
		UpdateExpression: aws.String("set Author = :a"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				S: aws.String(author),
			},
		},
	}

	// Perform the update
	_, err := svc.UpdateItem(updateInput)
	if err != nil {
		log.Fatalf("Got error calling UpdateItem for author update: %s", err)
	}

	log.Println("Successfully updated book's author")
}

func GetCurrentBook() Book {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	filt := expression.Name("Active").Equal(expression.Value(true))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		log.Fatalf("Got error building expression: %s", err)
	}

	booksTable := tableName("books")
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(booksTable),
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

	readingProgressTable := tableName("reading_progress")
	input := &dynamodb.PutItemInput{
		TableName: aws.String(readingProgressTable),
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

	readingProgressTable := tableName("reading_progress")
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(readingProgressTable),
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
		groupProgress += fmt.Sprintf("%s: %d%%\n", user.FullName, progress.Progress)
	}

	if groupProgress == "" {
		return "No users have set their progress yet."
	}

	return groupProgress
}

func GetUserDetails(userName string) User {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	usersTable := tableName("users")
	input := &dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
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
	sess := AWSsession()
	svc := dynamodb.New(sess)

	booksTable := tableName("books")
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(booksTable),
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
		FullName: name,
		IsAdmin:  false,
	})
	if err != nil {
		log.Fatalf("Failed to marshal User: %s", err)
	}

	usersTable := tableName("users")
	input := &dynamodb.PutItemInput{
		TableName: aws.String(usersTable),
		Item:      item,
	}

	// Put the item into the DynamoDB table
	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Failed to put User item into DynamoDB table: %s", err)
	}

	fmt.Printf("Successfully added user: %s\n", userName)
}

func SetUserFullName(name string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)
	usersTable := tableName("users")
	scanInput := &dynamodb.ScanInput{
		TableName:        aws.String(usersTable),
		FilterExpression: aws.String("attribute_not_exists(FullName) or attribute_type(FullName, :nullType)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":nullType": {S: aws.String("NULL")},
		},
	}

	// Perform the scan
	scanResult, err := svc.Scan(scanInput)
	fmt.Println("scanResult", scanResult)
	if err != nil {
		log.Fatalf("Failed to scan DynamoDB table: %s", err)
	}

	// Update each item found with the new name
	for _, item := range scanResult.Items {
		user := User{}
		err := dynamodbattribute.UnmarshalMap(item, &user)
		if err != nil {
			log.Fatalf("Failed to unmarshal DynamoDB item: %s", err)
		}

		// Set the new name
		user.FullName = name

		// Marshal the updated user back into a DynamoDB attribute map
		updatedAttributes, err := dynamodbattribute.MarshalMap(user)
		if err != nil {
			log.Fatalf("Failed to marshal updated user: %s", err)
		}

		// Create the input for the update
		updateInput := &dynamodb.PutItemInput{
			TableName: aws.String(usersTable),
			Item:      updatedAttributes,
		}

		_, err = svc.PutItem(updateInput)
		if err != nil {
			log.Fatalf("Failed to update item in DynamoDB table: %s", err)
		}

		fmt.Printf("Successfully updated user: %s\n", user.UserName)
	}
}

func RemoveUser(userName string) {
	sess := AWSsession()
	svc := dynamodb.New(sess)

	usersTable := tableName("users")
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(usersTable),
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

	booksTable := tableName("books")
	input := &dynamodb.ScanInput{
		TableName: aws.String(booksTable),
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

	usersTable := tableName("users")
	input := &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(usersTable),
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
	usersTable := tableName("users")
	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(usersTable),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItem(input)
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

	readingProgressTable := tableName("reading_progress")
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(readingProgressTable),
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
// 	SetUserFullName("111")
// }
