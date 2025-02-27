package main

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	userv1 "github.com/fraser-isbester/federated-gql/gen/go/user/v1"
)

func TestGetUser(t *testing.T) {
	server := &userServer{}
	ctx := context.Background()

	testCases := []struct {
		name          string
		userID        string
		expectedName  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "Get predefined user alice",
			userID:       "alice",
			expectedName: "Alice Johnson",
			expectError:  false,
		},
		{
			name:         "Get predefined user bob",
			userID:       "bob",
			expectedName: "Bob Smith",
			expectError:  false,
		},
		{
			name:         "Get dynamically generated user",
			userID:       "charlie",
			expectedName: "User charlie",
			expectError:  false,
		},
		{
			name:          "Empty user ID returns error",
			userID:        "",
			expectError:   true,
			errorContains: "user_id is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&userv1.GetUserRequest{
				UserId: tc.userID,
			})

			resp, err := server.GetUser(ctx, req)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', got no error", tc.errorContains)
				}
				if connectErr, ok := err.(*connect.Error); ok {
					if connectErr.Message() != tc.errorContains {
						t.Fatalf("expected error containing '%s', got '%s'", tc.errorContains, connectErr.Message())
					}
				} else {
					t.Fatalf("expected connect.Error, got %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Msg.User.Name != tc.expectedName {
				t.Errorf("expected user name '%s', got '%s'", tc.expectedName, resp.Msg.User.Name)
			}

			if resp.Msg.User.UserId != tc.userID {
				t.Errorf("expected user ID '%s', got '%s'", tc.userID, resp.Msg.User.UserId)
			}
		})
	}
}