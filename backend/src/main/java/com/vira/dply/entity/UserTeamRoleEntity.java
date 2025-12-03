package com.vira.dply.entity;

import java.time.Instant;
import java.util.UUID;

import com.vira.dply.enums.Role;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.Id;
import jakarta.persistence.Index;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(
  name = "user_team_roles",
  uniqueConstraints = {
    @UniqueConstraint(
      name = "uk_user_team_roles_user_team",
      columnNames = {"user_id", "team_id"}
    )
  },
  indexes = {
    @Index(name = "idx_user_team_roles_user", columnList = "user_id"),
    @Index(name = "idx_user_team_roles_team", columnList = "team_id")
  }
)
@Data
@NoArgsConstructor
@AllArgsConstructor
public class UserTeamRoleEntity {

  @Id
  @GeneratedValue
  private UUID id;

  @ManyToOne(optional = false, fetch = FetchType.LAZY)
  @JoinColumn(name = "user_id", nullable = false)
  private UserEntity user;

  @ManyToOne(optional = false, fetch = FetchType.LAZY)
  @JoinColumn(name = "team_id", nullable = false)
  private TeamEntity team;

  @Enumerated(EnumType.STRING)
  @Column(nullable = false, length = 50)
  private Role role;

  @Column(name = "created_at", nullable = false, updatable = false)
  private Instant createdAt = Instant.now();

}
