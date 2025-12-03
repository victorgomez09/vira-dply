package com.vira.dply.security;

import java.util.EnumSet;
import java.util.Set;

import org.springframework.stereotype.Component;

import com.vira.dply.enums.Role;

@Component
public class RolePermissions {

    public static final Set<Role> DEPLOY = EnumSet.of(Role.OWNER, Role.ADMIN, Role.DEVELOPER);

    public static final Set<Role> SCALE = EnumSet.of(Role.OWNER, Role.ADMIN);

    public static final Set<Role> VIEW = EnumSet.of(Role.OWNER, Role.ADMIN, Role.DEVELOPER, Role.VIEWER);

    public static void check(Role role, Set<Role> allowed) {
        if (!allowed.contains(role)) {
            throw new SecurityException("Forbidden for role: " + role);
        }
    }
}
