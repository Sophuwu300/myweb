package main

import "time"

type BlogPost struct {
	BlogID  string    `storm:"blogID"`
	Title   string    `storm:"title"`
	Date    time.Time `storm:"date"`
	Summary string    `storm:"summary"`
}
type BlogComment struct {
	CommentID string    `storm:"commentID"`
	BlogID    string    `storm:"blogID"`
	Comment   string    `storm:"comment"`
	Date      time.Time `storm:"date"`
	Name      string    `storm:"name"`
}
