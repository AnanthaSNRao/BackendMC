package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/myGo/simplebank/db/mock"
	db "github.com/myGo/simplebank/db/sqlc"
	"github.com/myGo/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestGetAccountApi(t *testing.T) {
	account := randomAccount()

	testcases := []struct {
		name          string
		accountID     int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
				requireBodyMatchAccount(t, recoder.Body, account)
			},
		},

		{
			name:      "NotFound",
			accountID: account.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recoder.Code)
			},
		},

		{
			name:      "InternalError",
			accountID: account.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recoder.Code)
			},
		},

		{
			name:      "InvalidID",
			accountID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
	}

	for i := range testcases {
		tc := testcases[i]

		t.Run(tc.name, func(t *testing.T) {
			crtl := gomock.NewController(t)

			defer crtl.Finish()

			store := mockdb.NewMockStore(crtl)
			tc.buildStub(store)

			server := NewServer(store)
			recoder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)

			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)
			server.router.ServeHTTP(recoder, request)

			tc.checkResponse(t, recoder)
		})

	}

}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
		Owner:    util.RandomOwner(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)

	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
