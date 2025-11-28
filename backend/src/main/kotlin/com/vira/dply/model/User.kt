package com.vira.dply.controller

import jakarta.persistence.Entity
import jakarta.persistence.Table
import jakarta.persistence.Id
import jakarta.persistence.GeneratedValue
import jakarta.persistence.GenerationType
import jakarta.persistence.SequenceGenerator
import jakarta.persistence.Column

@Entity
@Table(name = "users")
data class User (
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id", unique = true)
    var id: Long,

    @Email
    @Column(name = "email", nullable = false)
    var email: String,

    @Column(name = "full_name", nullable = false)
    var fullName: String, 

    @Column(name = "password", nullable = false)
    var password: String,

    @Column(name = "is_enabled")
    var isEnabled: Boolean = true,

    @Column(name = "role", nullable = false)
    var role: String = "USER",

    @Column(name = "created_at")
    var createdAt: Date = null,

    @Column(name = "updated_at")
    var updatedAt: Date = null
): Serializable