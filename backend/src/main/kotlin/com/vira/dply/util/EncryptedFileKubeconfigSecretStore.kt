package com.vira.dply.util

import org.springframework.stereotype.Component
import java.io.File
import java.security.SecureRandom
import javax.crypto.Cipher
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

@Component
class EncryptedFileKubeconfigSecretStore(
    private val baseDir: String = "/var/lib/paas/secrets",
    private val masterKeyProvider: MasterKeyProvider
) : KubeconfigSecretStore {

    init {
        File(baseDir).mkdirs()
    }

    override fun store(key: String, kubeconfigYaml: String): KubeconfigRef {
        val file = File(baseDir, "$key.enc")
        val encrypted = encrypt(kubeconfigYaml.toByteArray(), masterKeyProvider.key())
        file.writeBytes(encrypted)
        return KubeconfigRef("file", file.name)
    }

    override fun load(ref: String): String {
        val file = File(baseDir, ref.id)
        if (!file.exists()) throw IllegalStateException("Kubeconfig not found: ${ref.id}")
        val decrypted = decrypt(file.readBytes(), masterKeyProvider.key())
        return String(decrypted)
    }

    override fun delete(ref: KubeconfigRef) {
        val file = File(baseDir, ref.id)
        if (file.exists()) file.delete()
    }

    private fun encrypt(data: ByteArray, key: SecretKey): ByteArray {
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val iv = ByteArray(12)
        SecureRandom().nextBytes(iv)
        val spec = GCMParameterSpec(128, iv)
        cipher.init(Cipher.ENCRYPT_MODE, key, spec)
        val encrypted = cipher.doFinal(data)
        return iv + encrypted // prepend IV
    }

    private fun decrypt(data: ByteArray, key: SecretKey): ByteArray {
        val iv = data.copyOfRange(0, 12)
        val payload = data.copyOfRange(12, data.size)
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.DECRYPT_MODE, key, GCMParameterSpec(128, iv))
        return cipher.doFinal(payload)
    }
}