package com.vira.dply.repository;

import java.util.Optional;
import java.util.UUID;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.vira.dply.entity.EnvironmentEntity;

@Repository
public interface EnvironmentRepository extends JpaRepository<EnvironmentEntity, UUID> {
    Optional<EnvironmentEntity> findByName(String name);

    boolean existsByName(String name);
}