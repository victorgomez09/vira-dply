package com.vira.dply.model

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.FetchType
import jakarta.persistence.GeneratedValue
import jakarta.persistence.GenerationType
import jakarta.persistence.Id
import jakarta.persistence.JoinColumn
import jakarta.persistence.ManyToOne
import jakarta.persistence.Table
import jakarta.persistence.UniqueConstraint
import org.hibernate.annotations.CreationTimestamp
import java.time.Instant

@Entity
@Table(
    name = "environment_users",
    uniqueConstraints = [
        UniqueConstraint(columnNames = ["environment_id", "user_id"])
    ]
)
data class EnvironmentUser(
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id", unique = true)
    var id: Long,

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "environment_id", nullable = false)
    val environment: Environment,

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    val user: User,

    @Enumerated(EnumType.STRING)
    val role: EnvironmentRole = EnvironmentRole.VIEWER,

    @CreationTimestamp
    val createdAt: Instant = Instant.now()
)