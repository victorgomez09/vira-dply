package com.vira.dply.repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.vira.dply.entity.TeamEntity;
import com.vira.dply.entity.UserTeamRoleEntity;
import com.vira.dply.enums.Role;

@Repository
public interface UserTeamRoleRepository
        extends JpaRepository<UserTeamRoleEntity, UUID> {

    Optional<UserTeamRoleEntity> findByUser_IdAndTeam_Id(
            UUID userId,
            UUID teamId);

    boolean existsByUser_IdAndTeam_IdAndRoleIn(
            UUID userId,
            UUID teamId,
            Iterable<Role> roles);

        List<UserTeamRoleEntity> findByTeam(TeamEntity team);

        Optional<UserTeamRoleEntity> findByTeamIdAndUserId(UUID teamId, UUID userId);
}