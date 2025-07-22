import (

	// The JSON library will be useful for serializing go structs.
	"encoding/json"

	// Likewise, useful for debugging, etc.
	"encoding/hex"

	// The Datastore requires UUIDs to store key-value entries.
	"github.com/google/uuid"

	// Useful for debug messages, or string manipulation for datastore keys.
	"strings"

	// Want to import errors.
	"errors"

	// Optional. You can remove the "_" there, but please do not touch
	// anything else within the import bracket.
	_ "strconv"

)

// This serves two purposes:
// a) It shows you some useful primitives, and
// b) it suppresses warnings for items not being imported.
// Of course, this function can be deleted.
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("Key is %v, %v", pk, sk)
}

// Helper function: Takes the first 16 bytes and converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

func repeat(b []byte, count int) []byte {
	if count == 0 {
		return []byte{}
	}
	if count < 0 {
		panic("bytes: negative Repeat count")
	} else if len(b)*count/count != len(b) {
		panic("bytes: Repeat count causes overflow")
	}

	nb := make([]byte, len(b)*count)
	bp := copy(nb, b)
	for bp < len(nb) {
		copy(nb[bp:], nb[:bp])
		bp *= 2
	}
	return nb
}

func hasSuffix(s, suffix []byte) bool {
	return len(s) >= len(suffix) && string(s[len(s)-len(suffix):])==string(suffix)
}
func Repeat(b []byte, count int) []byte {
	if count == 0 {
		return []byte{}
	}
	// Since we cannot return an error on overflow,
	// we should panic if the repeat will generate
	// an overflow.
	// See Issue golang.org/issue/16237.
	if count < 0 {
		panic("bytes: negative Repeat count")
	} else if len(b)*count/count != len(b) {
		panic("bytes: Repeat count causes overflow")
	}

	nb := make([]byte, len(b)*count)
	bp := copy(nb, b)
	for bp < len(nb) {
		copy(nb[bp:], nb[:bp])
		bp *= 2
	}
	return nb
}

func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, errors.New("")
	}
	if b == nil || len(b) == 0 {
		return nil, errors.New("")
	}
	if len(b)%blocksize == 0 {
	    return b, nil
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], Repeat([]byte{byte(n)}, n))
	return pb, nil
}
func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, errors.New("")
	}
	if b == nil || len(b) == 0 {
		return nil, errors.New("")
	}
	if len(b)%blocksize != 0 {
		return nil, errors.New("")
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, errors.New("")
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, errors.New("")
		}
	}
	return b[:len(b)-n], nil
}


// User is the structure definition for a user record.
type User struct {
	Username string
	PrivKey []byte
	UUID uuid.UUID

	DSSKey userlib.DSSignKey
	PKEDec userlib.PKEDecKey
    Keys FileKeys
    Files map[string]uuid.UUID

}

type File struct {
    UUID uuid.UUID
    Contents []byte
}

type FileKeys struct {
    ReadKey map[string][]byte
    WriteKey map[string][]byte
}


// InitUser will be called a single time to initialize a new user.
func InitUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata

    // Username can't be empty, doesn't have a length limit?
    if username == "" || password == "" {
        return nil, errors.New("Either username or password was empty")
    }
    userdata.Keys.ReadKey = make(map[string][]byte)
    userdata.Keys.WriteKey = make(map[string][]byte)
    userdata.Files = make(map[string]uuid.UUID)

    userdata.Username = username
    userdata.PrivKey = userlib.Argon2Key([]byte(password), []byte(username), 16)
    macEval, err := userlib.HashKDF(userdata.PrivKey, []byte("verified"))
    macEval, err = userlib.HashKDF(userdata.PrivKey, []byte(username))
    userdata.UUID, err = uuid.FromBytes(macEval[:16])

    pkeEnc, pkeDec, err := userlib.PKEKeyGen()
    dssKey, dssVerify, err := userlib.DSKeyGen()
    userdata.PKEDec = pkeDec
    userdata.DSSKey = dssKey
    err = userlib.KeystoreSet(username+"_pk", pkeEnc)
    err = userlib.KeystoreSet(username+"_vk", dssVerify)

    formattedData, err := json.Marshal(userdata)
//    encryptedUser := userlib.SymEnc(userdata.PrivKey, userlib.RandomBytes(16), formattedData)

    authenticate, err := userlib.HMACEval(userdata.PrivKey, []byte("untampered"))
    userlib.DatastoreSet(userdata.UUID, append(formattedData, authenticate...))
    // generate AES key with which all files are encrypted
    // hmac on username to act as key to

	return &userdata, nil
}
func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata

	privKey := userlib.Argon2Key([]byte(password), []byte(username), 16)

	macEval, err := userlib.HashKDF(privKey, []byte("verified"))
    macEval, err = userlib.HashKDF(privKey, []byte(username))
    tryUUID, err := uuid.FromBytes(macEval[:16])

    rawData, auth := userlib.DatastoreGet(tryUUID)
    if(!auth) {
        return nil, errors.New("Username not found or incorrect password")
    }
    err = json.Unmarshal(rawData[:len(rawData)-64], &userdata)
    authenticate, err := userlib.HMACEval(privKey, []byte("untampered"))

    if !userlib.HMACEqual(authenticate, rawData[len(rawData)-64:]) {
        return nil, errors.New("User data tampered")
    }

	return userdataptr, nil
}

func (userdata *User) StoreFile(filename string, data []byte) (err error) {
    var filedata File
    var fKeys FileKeys

    // structure of file: length of contents, contents, followed by key structure
    fKeys.ReadKey = make(map[string][]byte)
    fKeys.WriteKey = make(map[string][]byte)

    privKData := userlib.RandomBytes(userlib.AESBlockSizeBytes)
    pubKData := userlib.Hash(privKData)[:userlib.AESBlockSizeBytes]

    pkeEnc, _ := userlib.KeystoreGet(userdata.Username+"_pk")
    ownerReadKey, _ := userlib.PKEEnc(pkeEnc, privKData)
    ownerWriteKey, _ := userlib.PKEEnc(pkeEnc, pubKData)

    fKeys.ReadKey[userdata.Username] = ownerReadKey
    fKeys.WriteKey[userdata.Username] = ownerWriteKey
    userdata.Keys.ReadKey[filename] = ownerReadKey
    userdata.Keys.WriteKey[filename] = ownerWriteKey

	filedata.UUID, _ = uuid.FromBytes([]byte(filename + userdata.Username)[:16])
	filedata.Contents = data

    formattedKeys, _ := json.Marshal(fKeys)
    formattedData, _ := json.Marshal(filedata)
    keyLen, _ := json.Marshal(uint32(4294967295 - len(formattedKeys)))

    paddedData, _ := pkcs7Pad(formattedData, userlib.AESBlockSizeBytes)
	formattedData = userlib.SymEnc(privKData, userlib.RandomBytes(userlib.AESBlockSizeBytes), paddedData)
    userlib.DebugMsg("StoreFile asdf %x", privKData)

    rawData := append(formattedData, formattedKeys...)
    compiledData := append(keyLen, rawData...)
	userlib.DatastoreSet(filedata.UUID, compiledData)
	userdata.Files[filename] = filedata.UUID
	return
}
func (userdata *User) LoadFile(filename string) (dataBytes []byte, err error) {
    var filedata File
    var fKeys FileKeys
	formattedFile, ok := userlib.DatastoreGet(userdata.Files[filename])
	if !ok {
    	return nil, errors.New(strings.ToTitle("File not found!"))
    }
    var keyLen uint32
    if len(formattedFile) < 10 {
        return nil, errors.New("file not found")
    }
    err = json.Unmarshal(formattedFile[:10], &keyLen)

    keyLen = 4294967295-keyLen
    if len(formattedFile)-int(keyLen) >= len(formattedFile) || len(formattedFile)-int(keyLen) < 10{
            return nil, errors.New("file not found")
    }
    formattedData := formattedFile[10:len(formattedFile)-int(keyLen)]
    formattedKeys := formattedFile[len(formattedFile)-int(keyLen):]
    err = json.Unmarshal(formattedKeys, &fKeys)
    userlib.DebugMsg("LoadFile test %x", formattedData)
    if _, ok := fKeys.ReadKey[userdata.Username]; !ok {
        return nil, errors.New("accessRevoked")
    }
    privDKey, err := userlib.PKEDec(userdata.PKEDec, fKeys.ReadKey[userdata.Username])
    if len(privDKey) != 16 || err != nil {
        return nil, errors.New("invalid key")
    }
    dataBytes = userlib.SymDec(privDKey, formattedData)
    formattedData, _ = pkcs7Unpad(dataBytes, userlib.AESBlockSizeBytes)
    _ = json.Unmarshal(formattedData, &filedata)
	return filedata.Contents, nil
	//End of toy implementation

}


func (userdata *User) AppendFile(filename string, data []byte) (err error) {
    var filedata File
    var fKeys FileKeys
 // Naive implementation :((
 //     filedata.UUID = userdata.Files[filename]
    ciphertext, _ := userlib.DatastoreGet(userdata.Files[filename])

    var keyLen uint32
    if len(ciphertext) < 10 {
        return errors.New("file not found")
    }
    _ = json.Unmarshal(ciphertext[:10], &keyLen)
    keyLen = 4294967295-keyLen
    formattedData := ciphertext[10:len(ciphertext)-int(keyLen)]
    formattedKeys := ciphertext[len(ciphertext)-int(keyLen):]
    _ = json.Unmarshal(formattedKeys, &fKeys)
    if _, ok := fKeys.ReadKey[userdata.Username]; !ok {
        return errors.New("accessRevoked")
    }
    userlib.DebugMsg("AppendFile1 %x", formattedData)

    privDKey, _ := userlib.PKEDec(userdata.PKEDec, fKeys.ReadKey[userdata.Username])
    userlib.DebugMsg("AppendFile2 %x", privDKey)
    lastBlock := userlib.SymDec(privDKey, formattedData)
    userlib.DebugMsg("AppendFile3 %x", lastBlock)
    lastBlock, _ = pkcs7Unpad(lastBlock, userlib.AESBlockSizeBytes)
    userlib.DebugMsg("AppendFile4 %x", lastBlock)
    err = json.Unmarshal(lastBlock, &filedata)
    userlib.DebugMsg("AppendFile5 %s", filedata.Contents)
    filedata.Contents = append(filedata.Contents, data...)
    userlib.DebugMsg("AppendFile6 %s", filedata.Contents)
    formattedData, _ = json.Marshal(filedata)
    formattedData, _ = pkcs7Pad(formattedData, userlib.AESBlockSizeBytes)
    formattedData = userlib.SymEnc(privDKey, userlib.RandomBytes(16), formattedData)
//     lastBlock := formattedData[len(formattedData)-userlib.AESBlockSizeBytes:]
//     _ = privDKey
//     lastBlock = userlib.SymDec(privDKey, lastBlock)
//     unpaddedFormatted, _ := pkcs7Unpad(lastBlock, userlib.AESBlockSizeBytes)
//     userlib.DebugMsg("AppendFile data %x, unpad %x", lastBlock, unpaddedFormatted)
//     unpaddedFormatted = append(unpaddedFormatted, data...)
//     paddedData, _ := pkcs7Pad(unpaddedFormatted, userlib.AESBlockSizeBytes)
//     userlib.DebugMsg("AppendFile data %x, unpad %x", unpaddedFormatted, paddedData)
//     newCiphertext := userlib.SymEnc(privDKey, userlib.RandomBytes(16), paddedData)
//       cipherLen, _ := json.Marshal(uint32(4294967295 - len(formattedKeys)))
//
//     updatedCiphertext := append(formattedData[:len(formattedData)-userlib.AESBlockSizeBytes], newCiphertext...) // trying to append to filecontents without unmarshalling
//     filedata.Contents = updatedCiphertext
//     updatedCiphertext, _ = json.Marshal(filedata)
//     userlib.DebugMsg("AppendFile dec %x, %x", formattedData, updatedCiphertext)
    rawData := append(formattedData, formattedKeys...)
    compiledData := append(ciphertext[:10], rawData...)

    userlib.DatastoreSet(userdata.Files[filename], compiledData)
	return
}
func (userdata *User) ShareFile(filename string, recipient string) (
	accessToken uuid.UUID, err error) {
    var fKeys FileKeys
    recPubKey, _ := userlib.KeystoreGet(recipient+"_pk")
    formattedFile, ok := userlib.DatastoreGet(userdata.Files[filename])
    if !ok {
        return userdata.Files[filename], errors.New("Check filename of file to be shared")
    }
    var keyLen uint32
    _ = json.Unmarshal(formattedFile[:10], &keyLen)
    keyLen = 4294967295-keyLen
    formattedData := formattedFile[10:len(formattedFile)-int(keyLen)]
    formattedKeys := formattedFile[len(formattedFile)-int(keyLen):]

    _ = json.Unmarshal(formattedKeys, &fKeys)

    privKBytes, _ := userlib.PKEDec(userdata.PKEDec, fKeys.ReadKey[userdata.Username])
    pubKBytes, _ := userlib.PKEDec(userdata.PKEDec, fKeys.WriteKey[userdata.Username])

    shareReadKey, _ := userlib.PKEEnc(recPubKey, privKBytes)
    shareWriteKey, _ := userlib.PKEEnc(recPubKey, pubKBytes)

    fKeys.ReadKey[recipient] = shareReadKey
    fKeys.WriteKey[recipient] = shareWriteKey

    formattedKeys, _ = json.Marshal(fKeys)
    formattedKeyLen, _ := json.Marshal(uint32(4294967295 - len(formattedKeys)))

    rawData := append(formattedData, formattedKeys...)
    compiledData := append(formattedKeyLen, rawData...)

    userlib.DatastoreSet(userdata.Files[filename], compiledData)

	return userdata.Files[filename], err
}
func (userdata *User) ReceiveFile(filename string, sender string,
accessToken uuid.UUID) error {
    var filedata File
    var fKeys FileKeys
    formattedFile, ok := userlib.DatastoreGet(accessToken)
     if !ok {
         return errors.New("Check filename of file to be received")
     }
    var keyLen int
    _ = json.Unmarshal(formattedFile[:10], &keyLen)
    keyLen = 4294967295-keyLen
    formattedData := formattedFile[10:len(formattedFile)-int(keyLen)]
    formattedKeys := formattedFile[len(formattedFile)-int(keyLen):]
    _ = json.Unmarshal(formattedKeys, &fKeys)

    privDKey, _ := userlib.PKEDec(userdata.PKEDec, fKeys.ReadKey[userdata.Username])
    if len(privDKey) != 16 {
        return errors.New("invalid key")
    }
    dataBytes := userlib.SymDec(privDKey, formattedData)
    formattedData, _ = pkcs7Unpad(dataBytes, userlib.AESBlockSizeBytes)
    _ = json.Unmarshal(formattedData, &filedata)

    userdata.Keys.ReadKey[filename] = fKeys.ReadKey[userdata.Username]
    userdata.Keys.WriteKey[filename] = fKeys.WriteKey[userdata.Username]
    userdata.Files[filename] = accessToken
	return nil
}
func (userdata *User) RevokeFile(filename string, targetUsername string) (err error) {
     var filedata File
     var fKeys FileKeys

    formattedFile, ok := userlib.DatastoreGet(userdata.Files[filename])
    if !ok {
     return errors.New("Check filename of file to be shared")
    }
    var keyLen uint32
    _ = json.Unmarshal(formattedFile[:10], &keyLen)
    keyLen = 4294967295-keyLen
    formattedData := formattedFile[10:len(formattedFile)-int(keyLen)]
    formattedKeys := formattedFile[len(formattedFile)-int(keyLen):]
    _ = json.Unmarshal(formattedKeys, &fKeys)

    privDKey, _ := userlib.PKEDec(userdata.PKEDec, fKeys.ReadKey[userdata.Username])
    formattedData = userlib.SymDec(privDKey, formattedData)
    _ = json.Unmarshal(formattedData, &filedata)

    delete(fKeys.ReadKey, targetUsername)
    delete(fKeys.WriteKey, targetUsername)

    privDKey = userlib.RandomBytes(userlib.AESBlockSizeBytes)
    pubDKey := userlib.Hash(privDKey)[:userlib.AESBlockSizeBytes]

    for user, _ := range fKeys.ReadKey {
        pkeEnc, _ := userlib.KeystoreGet(user+"_pk")
        newReadKey, _ := userlib.PKEEnc(pkeEnc, privDKey)
        newWriteKey, _ := userlib.PKEEnc(pkeEnc, pubDKey)
        fKeys.ReadKey[user] = newReadKey
        fKeys.ReadKey[user] = newWriteKey
    }

    formattedKeys, _ = json.Marshal(fKeys)
    lenFKeys, _ := json.Marshal(uint32(4294967295 - len(formattedKeys)))
    reformattedData := userlib.SymEnc(privDKey, userlib.RandomBytes(userlib.AESBlockSizeBytes), formattedData)

    rawData := append(reformattedData, formattedKeys...)
    compiledData := append(lenFKeys, rawData...)
	userlib.DatastoreSet(filedata.UUID, compiledData)
	return
}
