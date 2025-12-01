package com.vira.dply.util

import javax.crypto.SecretKey

interface MasterKeyProvider {
    fun key(): SecretKey
}
