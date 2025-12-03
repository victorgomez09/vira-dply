package com.vira.dply.util;

import org.springframework.stereotype.Component;

import com.vira.dply.enums.Role;

@Component
public class KubeRoleMapper {

    public String toKubeRole(Role role) {
        return switch (role) {
            case OWNER, ADMIN -> "admin"; // permiso completo en el namespace
            case VIEWER -> "view"; // solo lectura
            default -> "view";
        };
    }
}
