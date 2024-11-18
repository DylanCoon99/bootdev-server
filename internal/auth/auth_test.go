package auth


import (
	"time"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	//"github.com/golang-jwt/jwt/v5"
	"testing"
	"fmt"
)




func TestMakeJWT(t *testing.T) {

	// tests that the MakeJWT function returns a JWT string with no errors

	/*
		func signature:

			MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error)

	*/

	user_uuid, err := uuid.NewRandom()

	tokenSecret := "eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"

	expiresIn := time.Minute * 2

	token, err := MakeJWT(user_uuid, tokenSecret, expiresIn)

	//fmt.Println(token, err)

	if token == "" && err != nil {
		// MakeJWT failed
		fmt.Println(err.Error())
		t.Fatalf("Failed to make token")
	}

	fmt.Println(token)

}



func TestValidateJWT(t *testing.T) {

	// test the function that validates a JWT that has been instatiated

	user_uuid, err := uuid.NewRandom()

	tokenSecret := "eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"

	expiresIn := time.Minute * 2

	tokenString, err := MakeJWT(user_uuid, tokenSecret, expiresIn)

	user_uuid, err = ValidateJWT(tokenString, tokenSecret)

	empty_user_id, _ := uuid.Parse("")

	if user_uuid == empty_user_id || err != nil {
		fmt.Println(err.Error())
		t.Fatalf("Failed to validate token")
	}

	fmt.Println(user_uuid.String())
}


/*
func TestGetBearerToken(t *testing.T) {




}
*/