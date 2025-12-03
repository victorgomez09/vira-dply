package com.vira.dply.repository;

import java.util.Optional;
import java.util.UUID;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.vira.dply.entity.UserEntity;

@Repository
public interface UserRepository extends JpaRepository<UserEntity, UUID> {

    Boolean existsByEmail(String userEmail);

    Optional<UserEntity> findByEmail(String userEmail);
}
