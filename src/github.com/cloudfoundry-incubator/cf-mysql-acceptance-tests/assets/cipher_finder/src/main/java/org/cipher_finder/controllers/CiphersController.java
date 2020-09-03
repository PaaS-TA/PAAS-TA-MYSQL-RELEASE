package org.cipher_finder.controllers;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.HashMap;
import java.util.Map;

@RestController
public class CiphersController {

    @Autowired
    private DataSource dataSource;

    @RequestMapping("/ping")
    public String getPing() {
        return "OK";
    }
    @RequestMapping("/ciphers")
    public Map<String, String> getCipher() throws SQLException {
        final Connection connection = dataSource.getConnection();
        final Map<String, String> resultsAsMap = new HashMap<>();
        try {
            final Statement statement = connection.createStatement();
            final ResultSet resultSet = statement.executeQuery("SHOW STATUS LIKE 'ssl_cipher'");
            if (resultSet.next()) {
                final String cipher = resultSet.getString(2);
                resultsAsMap.put("cipher_used", cipher);
            }
            else {
                resultsAsMap.put("error", "No results from show status");
            }
        }
        finally {
            connection.close();
        }

        return resultsAsMap;
    }
}
