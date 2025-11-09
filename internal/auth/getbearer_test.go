package auth_test

import (
    "net/http"
    "testing"

    auth "github.com/frogonabike/chirpy/internal/auth"
)

func TestGetBearerToken(t *testing.T) {
    tests := []struct{
        name string
        header string
        setup func(h http.Header)
        want string
        wantErr bool
    }{
        {"missing", "", nil, "", true},
        {"wrong scheme", "Token abc", nil, "", true},
        {"empty token", "Bearer ", nil, "", true},
        {"valid", "Bearer token123", nil, "token123", false},
        {"lowercase scheme", "bearer token123", nil, "", true},
        {"extra spaces", "Bearer    token123", nil, "   token123", false},
        {"tab separator", "Bearer\ttoken123", nil, "", true},
        {"token with spaces", "Bearer token with spaces", nil, "token with spaces", false},
        {"multiple headers", "", func(h http.Header){ h.Add("Authorization","Bearer first"); h.Add("Authorization","Bearer second") }, "first", false},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            h := http.Header{}
            if tc.setup != nil {
                tc.setup(h)
            } else if tc.header != "" {
                h.Set("Authorization", tc.header)
            }
            got, err := auth.GetBearerToken(h)
            if (err != nil) != tc.wantErr {
                t.Fatalf("GetBearerToken error mismatch: got err=%v wantErr=%v", err, tc.wantErr)
            }
            if got != tc.want {
                t.Fatalf("GetBearerToken = %q, want %q", got, tc.want)
            }
        })
    }
}
