// Copyright 2015 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func Encrypt(key, text []byte) (ciphertext []byte, err error) {
	var blk cipher.Block
	if blk, err = aes.NewCipher(key); err != nil {
		return nil, err
	}
	ciphertext = make([]byte, aes.BlockSize+len(string(text)))
	vect := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, vect); err != nil {
		return
	}
	cfb := cipher.NewCFBEncrypter(blk, vect)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func Decrypt(key, ciphertext []byte) (text []byte, err error) {
	var blk cipher.Block
	if blk, err = aes.NewCipher(key); err != nil {
		return
	}
	if len(ciphertext) < aes.BlockSize {
		err = errors.New("ciphrtext too short")
		return
	}
	vect := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(blk, vect)
	cfb.XORKeyStream(ciphertext, ciphertext)
	text = ciphertext
	return
}
