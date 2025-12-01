package com.vira.dply.util

import org.springframework.stereotype.Component
import javax.crypto.SecretKey
import javax.crypto.spec.SecretKeySpec

@Component
class EnvMasterKeyProvider : MasterKeyProvider {
    override fun key(): SecretKey {
        val raw = System.getenv("PAAS_MASTER_KEY")
            ?: error("PAAS_MASTER_KEY env variable missing")
        val keyBytes = raw.toByteArray()
        if (keyBytes.size != 16 && keyBytes.size != 32) {
            throw IllegalArgumentException("Master key must be 16 or 32 bytes for AES")
        }
        return SecretKeySpec(keyBytes, "AES")
    }
}