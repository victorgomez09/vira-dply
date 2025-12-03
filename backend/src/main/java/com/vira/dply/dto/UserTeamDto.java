package com.vira.dply.dto;

import java.util.UUID;

import com.vira.dply.enums.TeamRole;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.RequiredArgsConstructor;

@Data
@AllArgsConstructor
@RequiredArgsConstructor
public class UserTeamDto {

    private UUID teamId;
    private UUID userId;
    private TeamRole role;
}
