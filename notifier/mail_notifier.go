/*Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/skyrings/skyring-common/conf"
	"github.com/skyrings/skyring-common/db"
	"github.com/skyrings/skyring-common/models"
	"github.com/skyrings/skyring-common/tools/logger"
	"gopkg.in/mgo.v2/bson"
	"net"
	"net/smtp"
	"strconv"
)

var client *smtp.Client

func setTLSMailClient(addr string, a smtp.Auth, skipVerify bool) error {
	c, err := smtp.Dial(addr)

	if err != nil {
		return err
	}
	host, _, _ := net.SplitHostPort(addr)
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: skipVerify,
		}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	client = c
	return nil
}

func setSSLMailClient(addr string, auth smtp.Auth, skipVerify bool) error {
	host, _, _ := net.SplitHostPort(addr)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: skipVerify,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err = c.Auth(auth); err != nil {
		return err
	}
	client = c
	return nil
}

func sendMail(from string, to []string, msg []byte) error {
	if err := client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func getNotifier() (models.MailNotifier, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	collection := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_MAIL_NOTIFIER)
	var notifier []models.MailNotifier
	if err := collection.Find(nil).All(&notifier); err != nil || len(notifier) == 0 {
		logger.Get().Warning("Unable to read MailNotifier from DB: %v", err)
		return notifier[0], err
	} else {
		return notifier[0], nil
	}
}

func getMailRecepients() ([]string, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_USER)
	var users []models.User
	var recepients []string
	if err := coll.Find(bson.M{"notificationenabled": true}).All(&users); err != nil {
		logger.Get().Critical("Could not retrieve the list of users from DB. Error: %v", err)
		return recepients, err
	}
	for _, user := range users {
		recepients = append(recepients, user.Email)
	}
	return recepients, nil
}

func SetMailClient(notifier models.MailNotifier) error {
	password, err := base64.StdEncoding.DecodeString(notifier.Passcode)
	auth := smtp.PlainAuth("", notifier.MailId, string(password), notifier.SmtpServer)
	hostPort := fmt.Sprintf("%s:%s", notifier.SmtpServer, strconv.Itoa(notifier.Port))
	if notifier.Encryption == "ssl" {
		err = setSSLMailClient(hostPort, auth, notifier.SkipVerify)
	} else {
		err = setTLSMailClient(hostPort, auth, notifier.SkipVerify)
	}
	if err != nil {
		logger.Get().Error("Error setting the Mail Client Error: %v", err)
		return errors.New(fmt.Sprintf("Error setting the Mail Client Error: %v", err))
	}
	return nil
}

func MailNotify(subject string, body string) error {
	notifier, err := getNotifier()
	if err != nil {
		logger.Get().Warning("Could not Get the notifier. Error: %v", err)
		return errors.New(fmt.Sprintf("Could not Get the notifier. Error: %v", err))
	}

	recepients, err := getMailRecepients()
	if err != nil || len(recepients) == 0 {
		logger.Get().Warning("Could not Get any recepients. Error: %v", err)
		return errors.New(fmt.Sprintf("Could not Get any recepients. Error: %v", err))
	}

	var to_list bytes.Buffer
	for _, id := range recepients {
		to_list.WriteString(id)
		to_list.WriteString(",")
	}

	msg := []byte("To: " + to_list.String() + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	if client == nil {
		err := SetMailClient(notifier)
		return err
	}

	err = sendMail(notifier.MailId, recepients, msg)
	if err != nil {
		// retry once again after setting the client, as client might have timed out
		if err := SetMailClient(notifier); err != nil {
			return err
		}
		if err = sendMail(notifier.MailId, recepients, msg); err != nil {
			logger.Get().Error("Could not Send the Mail Notification. Error: %v", err)
			return err
		}
	}
	return nil
}
