package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/rmatsuoka/apifs"
	"github.com/rmatsuoka/ya9p"
	"github.com/sivchari/gotwtr"
)

func unmarshalString(p []byte) (string, error) {
	return string(bytes.TrimSuffix(p, []byte("\n"))), nil
}

func userID(c *gotwtr.Client, username string) (string, error) {
	r, err := c.RetrieveSingleUserWithUserName(context.Background(), username)
	if err != nil {
		return "", err
	}
	return r.User.ID, nil
}

func main() {
	key := apifs.NewVal[string]("", unmarshalString)

	username := apifs.NewVal[string]("", unmarshalString)
	timeline := apifs.NewEvent(func() (io.Reader, error) {
		c := gotwtr.New(key.Get())
		id, err := userID(c, username.Get())
		if err != nil {
			return nil, err
		}
		b := new(bytes.Buffer)
		ts, err := c.UserTweetTimeline(context.Background(), id)
		if err != nil {
			return nil, err
		}
		for _, t := range ts.Tweets {
			fmt.Fprintf(b, "%q\n", t.Text)
			// fmt.Fprintln(b, strings.ReplaceAll(t.Text, "\n", "\\n"))
		}
		return b, nil
	})
	tdir := new(apifs.Dir)
	tdir.Mknod("username", username)
	tdir.Mknod("tweets", timeline)

	word := apifs.NewVal[string]("", unmarshalString)
	result := apifs.NewEvent(func() (io.Reader, error) {
		c := gotwtr.New(key.Get())
		ss, err := c.SearchRecentTweets(context.Background(), word.Get())
		if err != nil {
			return nil, err
		}
		b := new(bytes.Buffer)
		for _, t := range ss.Tweets {
			fmt.Fprintf(b, "%q\n", t.Text)
			// fmt.Fprintln(b, strings.ReplaceAll(t.Text, "\n", "\\n"))
		}
		return b, nil
	})
	sdir := new(apifs.Dir)
	sdir.Mknod("word", word)
	sdir.Mknod("result", result)

	root := new(apifs.Dir)
	root.Mknod("key", key)
	root.Mknod("timeline", tdir)
	root.Mknod("search", sdir)
	fsys := apifs.NewFS(root)

	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
		}
		go ya9p.ServeFS(conn, fsys)
	}
}
