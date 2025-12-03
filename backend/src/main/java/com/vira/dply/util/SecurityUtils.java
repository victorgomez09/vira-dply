package com.vira.dply.util;

import java.util.UUID;

import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Component;

import com.vira.dply.security.JwtUserPrincipal;

@Component
public final class SecurityUtils {

  public static UUID currentUserId() {
    JwtUserPrincipal p = (JwtUserPrincipal) SecurityContextHolder
        .getContext()
        .getAuthentication()
        .getPrincipal();

    return p.userId();
  }
}