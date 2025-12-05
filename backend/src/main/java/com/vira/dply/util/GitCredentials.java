package com.vira.dply.util;

public record GitCredentials(
    String username,        // opcional
    String passwordOrToken, // token HTTPS
    String privateKey,      // SSH
    String passphrase       // SSH
) {}
