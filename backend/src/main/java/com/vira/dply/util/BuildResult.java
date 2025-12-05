package com.vira.dply.util;

import com.vira.dply.enums.BuildStatus;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class BuildResult {
    private BuildStatus status;
    private String logs;
    private String imageTag;
}
