package data

import (
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
)

func GetKeys() []transaction.PubKeyHash {
	return []transaction.PubKeyHash{
		//Address 1:
		// {
		//	"newVSAddress": "18oqBFQTkJkvzettLrmBK9xMgLpJ4F4TBT",
		//	"newPrivateKeyWIF": "5KWJj48nCa8Qj3wRfamGUCJ9sdPQzsJVHMUKSn8ynhXr2vMd9pB"
		// }
		[20]byte{85, 164, 59, 9, 50, 192, 165, 51, 248, 195, 88, 192, 248, 224, 36, 48, 107, 41, 229, 213},
		// Address 2:
		//{
		//  "newVSAddress": "1EseydVHgu7ap44xi9jReSJLVmuNPyrKWT",
		//  "newPrivateKeyWIF": "5JV9UxeFN6aTcgneVz6JkWQsWYLTNSmgGvrAfDydFEkBNwp3Zw3"
		//}
		[20]byte{152, 46, 34, 100, 204, 231, 171, 65, 176, 5, 125, 164, 176, 142, 37, 110, 207, 126, 19, 139},
		//Address 3:
		//{
		//  "newVSAddress": "16FBR3LcbXSposvNbTAj8cVsBF6PK9LACE",
		//  "newPrivateKeyWIF": "5JqSNQ7fRnf99ShSNBEirhfoXhYkpwy5wSKgYbfZzfkRWUR2tJe"
		//}
		[20]byte{57, 135, 37, 93, 176, 171, 67, 17, 38, 245, 25, 161, 186, 166, 42, 125, 33, 252, 160, 60},
		//Address 4:
		//{
		//  "newVSAddress": "14vwWSXoFMHnkm75RpdvjpoSLB7EGLGTPA",
		//  "newPrivateKeyWIF": "5Hze5mBjohHTiGmo1pEngz9msyssh9bFDjyUSwfQrF51h7bducf"
		//}
		[20]byte{43, 27, 236, 137, 83, 253, 119, 206, 236, 3, 191, 245, 225, 7, 223, 231, 233, 176, 191, 33},
	}
}
