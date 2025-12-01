package com.vira.dply.model

import jakarta.persistence.*
import jakarta.validation.constraints.Email
import org.springframework.security.core.GrantedAuthority
import org.springframework.security.core.authority.SimpleGrantedAuthority
import org.springframework.security.core.userdetails.UserDetails
import java.io.Serializable
import java.util.*

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
    var encodedPassword: String,

    @Column(name = "roles")
    @Enumerated(EnumType.STRING)
    var roles: List<Role>,

    @Column(name = "created_at")
    var createdAt: Date,

    @Column(name = "updated_at")
    var updatedAt: Date
): Serializable, UserDetails {

    override fun getAuthorities(): Collection<GrantedAuthority> {
        return roles.stream().map { authority: Role -> SimpleGrantedAuthority(authority.toString()) }.toList()
    }

    override fun getPassword(): String {
        return encodedPassword
    }

    override fun getUsername(): String {
        return email
    }

    override fun isAccountNonExpired(): Boolean {
        return true
    }

    override fun isAccountNonLocked(): Boolean {
        return true
    }

    override fun isCredentialsNonExpired(): Boolean {
        return true
    }

    override fun isEnabled(): Boolean {
        return true
    }
}