package com.vira.dply.repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.vira.dply.entity.EnvironmentEntity;

@Repository
public interface EnvironmentRepository extends JpaRepository<EnvironmentEntity, UUID> {
    Optional<EnvironmentEntity> findByName(String name);

    boolean existsByName(String name);

    @Query("""
        select distinct e
        from EnvironmentEntity e
        join e.teams t
        join t.userRoles utr
        where utr.user.id = :userId
    """)
    List<EnvironmentEntity> findAllAccessibleByUser(@Param("userId") UUID userId);
}