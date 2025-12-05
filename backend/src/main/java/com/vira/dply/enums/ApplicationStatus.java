package com.vira.dply.enums;

public enum ApplicationStatus {
    CREATED,     // Alta realizada
    BUILDING,    // Nixpacks corriendo
    DEPLOYING,   // Kubernetes
    RUNNING,     // OK
    FAILED       // Algo falló
}