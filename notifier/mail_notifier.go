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
	"github.com/skyrings/skyring-common/dbprovider"
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

func getNotifier(ctxt string, dbProvider dbprovider.DbInterface) (models.MailNotifier, error) {
	var notifier models.MailNotifier
	notifier, err := dbProvider.MailNotifierInterface().MailNotifier(ctxt)
	if err != nil {
		logger.Get().Warning("%s-Unable to read MailNotifier from DB: %v", ctxt, err)
		return notifier, err
	} else {
		return notifier, nil
	}
}

func getMailRecepients(ctxt string, dbProvider dbprovider.DbInterface) ([]string, error) {
	var users []models.User
	var recepients []string
	users, err := dbProvider.UserInterface().Users(bson.M{"notificationenabled": true})
	if err != nil {
		logger.Get().Critical(fmt.Sprintf("%s-Could not retrieve the list of users from DB. Error: %v", ctxt, err))
		return recepients, err
	}
	for _, user := range users {
		recepients = append(recepients, user.Email)
	}
	return recepients, nil
}

func SetMailClient(notifier models.MailNotifier, ctxt string) error {
	password, err := base64.StdEncoding.DecodeString(notifier.Passcode)
	auth := smtp.PlainAuth("", notifier.MailId, string(password), notifier.SmtpServer)
	hostPort := fmt.Sprintf("%s:%s", notifier.SmtpServer, strconv.Itoa(notifier.Port))
	if notifier.Encryption == "ssl" {
		err = setSSLMailClient(hostPort, auth, notifier.SkipVerify)
	} else {
		err = setTLSMailClient(hostPort, auth, notifier.SkipVerify)
	}
	if err != nil {
		logger.Get().Error("%s-Error setting the Mail Client Error: %v", ctxt, err)
		return errors.New(fmt.Sprintf("Error setting the Mail Client Error: %v", err))
	}
	return nil
}

func MailNotify(subject string, body string, dbProvider dbprovider.DbInterface, ctxt string) error {
	notifier, err := getNotifier(ctxt, dbProvider)
	if err != nil {
		logger.Get().Warning("%s-Could not Get the notifier. Error: %v", ctxt, err)
		return errors.New(fmt.Sprintf("Could not Get the notifier. Error: %v", err))
	}

	recepients, err := getMailRecepients(ctxt, dbProvider)
	if err != nil || len(recepients) == 0 {
		logger.Get().Warning("%s-Could not Get any recepients. Error: %v", ctxt, err)
		return errors.New(fmt.Sprintf("Could not Get any recepients. Error: %v", err))
	}

	var to_list bytes.Buffer
	for _, id := range recepients {
		to_list.WriteString(id)
		to_list.WriteString(",")
	}
	subject = fmt.Sprintf("%s: ", notifier.SubPrefix) + subject
	msg := []byte("To: " + to_list.String() + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	if client == nil {
		err := SetMailClient(notifier, ctxt)
		if err != nil {
			return err
		}
	}
	if !notifier.MailNotification {
		return nil
	}
	err = sendMail(notifier.MailId, recepients, msg)
	if err != nil {
		// retry once again after setting the client, as client might have timed out
		if err := SetMailClient(notifier, ctxt); err != nil {
			if err != nil {
				return err
			}
		}
		if err = sendMail(notifier.MailId, recepients, msg); err != nil {
			logger.Get().Error("%s-Could not Send the Mail Notification. Error: %v", ctxt, err)
			return err
		}
	}
	return nil
}

func TestMailNotify(notifier models.MailNotifier, subject string, body string, recepient []string, ctxt string) error {
	subject = fmt.Sprintf("%s: ", notifier.SubPrefix) + subject
	msg := []byte("To: " + recepient[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	err := SetMailClient(notifier, ctxt)
	if err != nil {
		logger.Get().Error("%s-Error setting the Mail Client Error: %v", ctxt, err)
		return err
	}

	err = sendMail(notifier.MailId, recepient, msg)
	if err != nil {
		// retry once again after setting the client, as client might have timed out
		if err := SetMailClient(notifier, ctxt); err != nil {
			if err != nil {
				logger.Get().Error("%s-Error setting the Mail Client Error: %v", ctxt, err)
				return err
			}
		}
		if err := sendMail(notifier.MailId, recepient, msg); err != nil {
			logger.Get().Error("%s-Could not Send the Mail Notification. Error: %v", ctxt, err)
			return err
		}
	}
	client = nil
	return nil
}
