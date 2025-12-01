package com.vira.dply.model

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.GeneratedValue
import jakarta.persistence.GenerationType
import jakarta.persistence.Id
import jakarta.persistence.Table
import jakarta.persistence.UniqueConstraint
import org.hibernate.annotations.CreationTimestamp
import org.hibernate.annotations.UpdateTimestamp
import java.time.Instant

@Entity
@Table(
    name = "environments", uniqueConstraints = [
        UniqueConstraint(columnNames = ["name"])
    ]
)
data class Environment(
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id", unique = true)
    var id: Long,

    @Column(nullable = false)
    var name: String,               // dev, staging, prod, local-1

    @Enumerated(EnumType.STRING)
    var status: EnvironmentStatus = EnvironmentStatus.CREATING,

    var description: String?,
    var kubeconfigRef: String?,

    @CreationTimestamp
    val createdAt: Instant = Instant.now(),

    @UpdateTimestamp
    val updatedAt: Instant = Instant.now()
)