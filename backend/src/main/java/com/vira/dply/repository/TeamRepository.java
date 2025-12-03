package com.vira.dply.repository;

import java.util.Optional;
import java.util.UUID;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.vira.dply.entity.TeamEntity;

@Repository
public interface TeamRepository extends JpaRepository<TeamEntity, UUID> {

    Optional<TeamEntity> findByNameAndEnvironment_Id(String name, UUID environmentId);

    boolean existsByNameAndEnvironment_Id(String name, UUID environmentId);
}