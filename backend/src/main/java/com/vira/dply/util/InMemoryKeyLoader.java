package com.vira.dply.util;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.security.GeneralSecurityException;
import java.security.KeyPair;
import java.util.Collection;

import org.apache.sshd.common.NamedResource;
import org.apache.sshd.common.config.keys.FilePasswordProvider;
import org.apache.sshd.common.util.security.SecurityUtils;
import org.springframework.stereotype.Component;

@Component
public class InMemoryKeyLoader {
    /**
     * Carga un Iterable<KeyPair> a partir de la clave privada en memoria
     * @param privateKey PEM string
     * @param passphrase passphrase o null si no tiene
     * @return Iterable<KeyPair>
     */
    public Iterable<KeyPair> load(String privateKey, String passphrase) {
        try {
            FilePasswordProvider passwordProvider =
                    passphrase == null ? FilePasswordProvider.EMPTY
                            : (session, resource, i) -> passphrase;

            Collection<KeyPair> keys = (Collection<KeyPair>) SecurityUtils.loadKeyPairIdentities(
                    null,  // SessionContext no necesario
                    NamedResource.ofName("in-memory-key"),
                    new ByteArrayInputStream(privateKey.getBytes()),
                    passwordProvider
            );

            return keys;

        } catch (IOException | GeneralSecurityException e) {
            throw new IllegalArgumentException("Failed to load SSH private key", e);
        }
    }
}
