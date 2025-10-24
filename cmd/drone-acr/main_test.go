package main

import (
    "os"
    "testing"
)

func TestGetAuthInputValidation(t *testing.T) {
    // missing tenant
    if _, _, err := getAuth("client", "secret", "", "", "sub", "registry.azurecr.io"); err == nil {
        t.Fatalf("expected error for missing tenantId")
    }
    // missing clientId
    if _, _, err := getAuth("", "secret", "", "tenant", "sub", "registry.azurecr.io"); err == nil {
        t.Fatalf("expected error for missing clientId")
    }
    // missing both secret and cert
    if _, _, err := getAuth("client", "", "", "tenant", "sub", "registry.azurecr.io"); err == nil {
        t.Fatalf("expected error for missing credentials")
    }
}

func TestGetenvAuthorityHost(t *testing.T) {
    os.Setenv("AZURE_AUTHORITY_HOST", "https://login.microsoftonline.us")
    defer os.Unsetenv("AZURE_AUTHORITY_HOST")

    got := getenv("AZURE_AUTHORITY_HOST")
    if got != "https://login.microsoftonline.us" {
        t.Fatalf("expected AZURE_AUTHORITY_HOST to be returned, got %q", got)
    }
}

