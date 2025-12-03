package com.vira.dply.entity;

import java.util.HashSet;
import java.util.Set;
import java.util.UUID;

import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.Id;
import jakarta.persistence.OneToMany;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "environments", uniqueConstraints = {
    @UniqueConstraint(name = "uk_environment_name", columnNames = {"name"})
})
@Data
@AllArgsConstructor
@NoArgsConstructor
public class EnvironmentEntity {

    @Id
    @GeneratedValue
    private UUID id;

    @Column(nullable = false, unique = true)
    private String name;

    @Column(nullable = false)
    private String kubeContext; // nombre del contexto de Kubernetes

    @Column(name = "kube_config_path", nullable = false)
    private String kubeConfigPath;

    @OneToMany(
        mappedBy = "environment",
        cascade = CascadeType.ALL,
        orphanRemoval = true
    )
    private Set<TeamEntity> teams = new HashSet<>();
}