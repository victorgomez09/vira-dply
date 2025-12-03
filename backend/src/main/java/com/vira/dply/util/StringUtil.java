package com.vira.dply.util;

import org.springframework.stereotype.Component;

@Component
public class StringUtil {
    public String removeTrailingDash(String value) {
        if (value != null && value.endsWith("-")) {
            return value.substring(0, value.length() - 1);
        }
        return value;
    }
}
