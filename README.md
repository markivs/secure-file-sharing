# secure-file-sharing
Applied cryptographic primitives in creating end-to-end encrypted file sharing system with efficient sharing, appending, and revoking capabilities. Designing a true, trusted E2EE system is not the goal of this project, but I'm using it to document and perform a security analysis exercise. The Golang code written interfaces with APIs from a third-party file server: the goal is to secure the system against those with direct access to files on the server.

# System Design
## How is a file stored on the server?
The design requirements require using a scheme which user sessions can use to verify identity. Keeping in mind the need for many different users using the application at the same time and single users having multiple active sessions at the same time (no explicit need to maintain independence between user sessions), I’ve decided on the following attributes for a given user, upon which the file system is established. A given user will keep track of their username, private 16-byte symmetric key, public UUID, and RSA decryption key and signature key. For the purposes of file storage, an additional FileKey structure (denoted below) and map of personal file UUIDs are required.

A given file will be encrypted with AES-CBC (IV randomly generated per instance) to theoretically allow for efficient append operations. Each file will have its own data-key pair, separate from the user-key pair.  The public key from the pair is used to encrypt the plaintext. The resultant ciphertext is stored on the server. This mechanism of a separate data-key pair is used in future sharing mechanisms.

To begin, all users have user-key pairs involving a private key for decryption and a corresponding public key which other users will use to create ciphertext such that only the holder of the private key can access the plaintext.

To store a file securely, a user will generate a data-key pair for the file and immediately encrypt the plaintext with the ‘public’ data-key component. Next, a read-key and write-key is generated for each user who is given access, including the owner.  These will be sent in a plaintext “FileKey” structure along with the file ciphertext structure.
The read key is generated by encrypting the ‘private’ data-key component with the accessing users’ public key [PKEEnc(pubUserKey, privDataKey)]. Likewise, the write key is generate by encrypting the ‘public’ data-key component with the accessing users’ public key [PKEEnc(pubUserKey, pubDataKey)]. Since only the holder of the corresponding privUserKey can access either the ‘private’ or ‘public’ data-key component, the original data-key pair is secured.

Finally, for ease of separating the individual structs once they have been marshalled, the size of the marshalled FileKeys are provided at the start of the bytestream.
The final structure:
Bytestream Component
Position (byte slices)
uint32 fileKeyLen
[ : 10]
marshalled, encrypted, padded file struct
[10 :len(bytestream)-FileKeyLen-len(FileKeyLen)-len(HMAC)]
plaintext FileKey struct containing all accessing users’ read and write keys
[len(bytestream)-FileKeyLen-len(FileKeyLen)-len(HMAC : len(bytestream)-64]
HMAC
[len(bytestream)-64 : ]

During the decryption/loading process, the fileKeyLen is used to separate and unmarshal the plaintext FileKey struct to extract the read and write keys. The user uses their private RSA key component to decrypt the read key, allowing access to the file’s private data-key component. This private data-key component gives access to the plaintext.

An HMAC is generated for the purpose of checking for tampering. This HMAC is generated using the public component of the data-key pair, accessible to all validated users. Thus, all collaborators on the file can change the file without flagging for tampering.

## How does a file get shared with another user?

Keystore has a registry of user-key pairs where the public key of the user to be shared with can be accessed. To allow multiple users to read the encrypted file, the following mechanism is used.
Encrypt the private key from the data-key pair with the sharee’s user-key public key. The resultant read-key is appended to the preexisting structure and marshalled.
Encrypt the public key from the data-key pair with the sharee’s user-key public key. The resultant write-key is appended to the preexisting structure and marshalled. (Owner and all sharee’s each have their own write-key) 
The resultant byte stream is re-shared with the server. 
The server is not in possession of any of the shared users’ private keys, the server can’t decrypt the file’s data-key pair and thus can’t gain access to the plaintext.
What is the process of revoking a user’s access to a file?
Revoking access requires re-encryption of the entire file, and that the user revoking access has the original private component of the data-key pair (the plaintext was encrypted with the public component). The revoker obtains the plaintext by decrypting with the private key.
With the plaintext, the revoker first generates a new data-key pair with a private key and corresponding public key. The plaintext is encrypted using the new public component. They then remove the revokee’s preexisting read and write keys from the structure. The remaining users are generated new read and write keys using the aforementioned readKey = [PKEEnc(pubUserKey, privDataKey)], writeKey = [PKEEnc(pubUserKey, pubDataKey)]. The revokee’s previous read key and write key can no longer be used to access the current version of the file on the datastore.

## How does the design support efficient file append?

The user begins by decrypting their write-key with their private key to obtain the public data-key. 

The file structure specifies that the file contents are padded, encrypted, then marshalled (in that order). In line with this format, and keeping in mind that the file contents are to be largely untouched to support efficient file append, the scheme uses the last block of the encrypted data. The last block (of static length specified in AESBlockSizeBytes) is decrypted individually, then unpadded, to obtain unformatted plaintext of an unspecified size.

The appending data is combined with the freshly-obtained chunk of plaintext, then re-padded and re-encrypted using the public data-key obtained from the write-key. Then, this value can simply be appended to the end of the original, untouched ciphertext. This process should be carried out while ignoring the last block of the original ciphertext, which has been integrated into the new ciphertext. The resultant ciphertext can be updated to the file struct, then sent to the datastore following aforementioned procedures (re-appending HMAC, etc). 

This method does not grow in complexity with increasing lengths of either original file size or new file size. 

# Security Analysis
## Attack 1: Rollback
The symmetric encryption algorithm used on the file struct makes use of IVs to counteract the potential for replay attacks. The use of an IV necessitates that encrypted files are serially dissimilar and tied to the time they were encrypted. In addition, the verification and tamper algorithm is tied to the timing of byte stream formation, so the chances of replay causing an unintended interaction are insignificant. 
A valid edit to the file requires a valid HMAC using the files public data-key component. Utilizing an authentication code generated by a revoked user will not have been generated by the current public component of the new file encryption, thus invalidating any rollback that happens to occur despite other measures inherent to the encryption algorithm. This revoking method also means all shares from a revoked user from any point in time also become invalid, further cementing the futility of this attack. As a whole, all of these components make revoked adversaries unlikely.
## Attack 2: Modification
In both user struct and file struct sharing processes, the encrypted byte stream is tied to an HMAC generated for the sole purpose of identifying that any changes to the file were intended by a user with access to a unique write-key. This means both revoked and datastore adversaries cannot tamper with the MAC associated filedata without detection. The user struct is especially of note in this process, generating HMAC based on a child hash of the the user’s private key which additionally denies tampering with the Init/GetUser functionality
## Attack 3: MITM 
An adversary with the ability to execute a MITM attack could potentially alter the UUID of a shared file, which could result in undesired malicious behavior. The UUID is thus contained within the file struct and can be additionally verified by unmarshalling the structure upon successful file receival. Successful file receival is ensured by the aforementioned message authentication procedures, so tampering with radical components of the byte stream is not computationally realistic.
