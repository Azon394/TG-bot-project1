#include <iostream>
#include <string>
#include "include/httplib.h"
#include "httplib.h"
#include "nlohmann/json.hpp"
#include <urlmon.h>
#pragma comment(lib, "urlmon.lib")
#include "xlnt/xlnt.hpp"
#include <Windows.h>
#include <ctype.h>
#include <vector>
#include <queue>
#include <jwt-cpp/jwt.h>
#include <fstream>
#include <chrono>
#include <thread>
#include <functional>
using namespace httplib;
using json = nlohmann::json;
using size_t = std::size_t;
using namespace httplib;

void parsing() {
        SetConsoleOutputCP(65001);
        xlnt::workbook wb;
        xlnt::worksheet ws;
        wb.load("C:\\Scripts\\file.xlsx");
        json json_data;        
        ws = wb.sheet_by_index(1);

        auto rows_count = ws.calculate_dimension().height();
        auto cols_count = ws.calculate_dimension().width();

        std::vector <std::pair <size_t, size_t>> weeks;
        std::vector <std::pair <size_t, size_t>> groups;

        for (size_t col = 1; col <= cols_count; col++) {
            for (size_t row = 1; row <= rows_count; row++) {
                if (ws.cell(col, row).to_string().find("группа") != std::string::npos) {
                    groups.push_back(std::make_pair(col, row));
                }
                if (ws.cell(col, row).to_string().find("неделя") != std::string::npos) {
                    weeks.push_back(std::make_pair(col, row));
                }
            }
        }

        std::map<std::string, std::pair<size_t, size_t>> practice_lessons; //ЛР ПЗ ЛК - все существующие виды занятий

        size_t cnt = 0; //количество групп, которые пропарсили
        for (auto& coordinates : groups) {
            std::vector<json> curr_days;
            cnt += 1;
            size_t x = coordinates.first;
            size_t y = coordinates.second;
            size_t lesson = y + 2;      // x; y+2
            size_t lesson_num_x;
            size_t lesson_num_y;
            if (cnt <= (groups.size() / 2)) {
                lesson_num_x = weeks[0].first + 1;  // координаты x по неделе нечетной
                lesson_num_y = lesson - 1;          // координаты y по неделе нечетной
            }
            else {
                lesson_num_x = weeks[1].first + 1;  // координаты x по неделе четной
                lesson_num_y = lesson - 1;          // координаты y по неделе четной
            }
            size_t name_x = lesson_num_x - 1;
            size_t name_y = lesson_num_y;
            size_t curr_num;
            int prev_num = -100;
            if (ws.cell(lesson_num_x, lesson_num_y).has_value()) {
                curr_num = stoi(ws.cell(lesson_num_x, lesson_num_y).to_string());
            }
            else {
                curr_num = prev_num;
            }
            std::vector<json> curr_lessons;
            std::queue<std::string> dates;
            while (lesson_num_y <= rows_count) {
                if (prev_num > curr_num && ws.cell(name_x, name_y).to_string() != "") {
                    dates.push(ws.cell(name_x, name_y).to_string());
                    if (!curr_lessons.empty()) {
                        curr_days.push_back({
                            {dates.front() , curr_lessons }
                            });
                        dates.pop();
                    }
                    curr_lessons = {};
                    if (ws.cell(x, name_y).is_merged()) {
                        curr_lessons.push_back({
                        { "Название урока", ws.cell(x, name_y).to_string() },
                        { "Преподаватель", ws.cell(x, name_y + 1).to_string() },
                        { "Местонахождение", ws.cell(x, name_y + 2).to_string() },
                        { "Тип занятия", ws.cell(x - 1, name_y).to_string() },
                        { "Номер занятия", stoi(ws.cell(lesson_num_x, name_y).to_string()) },
                        { "Комментарий", ""},
                        { "Подгруппа", 0}
                            });
                    }
                    else {
                        size_t k = 0;
                        while ((ws.cell(x + k, y).to_string() != "вид занятий") && (x + k <= cols_count)) {
                            if (ws.cell(x + k, name_y).has_value()) {
                                curr_lessons.push_back({
                                { "Название урока", ws.cell(x + k, name_y).to_string() },
                                { "Преподаватель", ws.cell(x + k, name_y + 1).to_string() },
                                { "Местонахождение", ws.cell(x + k, name_y + 2).to_string() },
                                { "Тип занятия", ws.cell(x - 1, name_y).to_string() },
                                { "Номер занятия", stoi(ws.cell(lesson_num_x, name_y).to_string()) },
                                { "Комментарий", ""},
                                { "Подгруппа",k + 1}
                                    });
                            }
                            k++;
                        }
                    }
                }
                if (prev_num < curr_num) {
                    if (ws.cell(x, name_y).is_merged()) {
                        curr_lessons.push_back({
                        { "Название урока", ws.cell(x, name_y).to_string() },
                        { "Преподаватель", ws.cell(x, name_y + 1).to_string() },
                        { "Местонахождение", ws.cell(x, name_y + 2).to_string() },
                        { "Тип занятия", ws.cell(x - 1, name_y).to_string() },
                        { "Номер занятия", stoi(ws.cell(lesson_num_x, name_y).to_string()) },
                        { "Комментарий", ""},
                        { "Подгруппа", 0}
                            });
                    }
                    else {
                        size_t k = 0;
                        while ((ws.cell(x + k, y).to_string() != "вид занятий") && (x + k <= cols_count)) {
                            if (ws.cell(x + k, name_y).has_value()) {
                                curr_lessons.push_back({
                                { "Название урока", ws.cell(x + k, name_y).to_string() },
                                { "Преподаватель", ws.cell(x + k, name_y + 1).to_string() },
                                { "Местонахождение", ws.cell(x + k, name_y + 2).to_string() },
                                { "Тип занятия", ws.cell(x - 1, name_y).to_string() },
                                { "Номер занятия", stoi(ws.cell(lesson_num_x, name_y).to_string()) },
                                { "Комментарий", ""},
                                { "Подгруппа",k + 1}
                                    });
                            }
                            k++;
                        }
                    }
                }
                lesson_num_y++;
                if (ws.cell(lesson_num_x, lesson_num_y).has_value()) {
                    prev_num = curr_num;
                    curr_num = stoi(ws.cell(lesson_num_x, lesson_num_y).to_string());
                    name_y = lesson_num_y;
                }
                else {
                    prev_num = curr_num;
                }
            }
            dates.push(ws.cell(name_x, name_y).to_string());
            curr_days.push_back({
                {dates.front() , curr_lessons}
                });
            dates.pop();
            if (cnt <= (groups.size() / 2)) {
                json_data[ws.cell(weeks[0].first, weeks[0].second + 1).to_string()][ws.cell(x, y).to_string()][ws.cell(weeks[0].first, weeks[0].second).to_string()] = curr_days;
            }
            else {
                json_data[ws.cell(weeks[1].first, weeks[1].second + 1).to_string()][ws.cell(x, y).to_string()][ws.cell(weeks[1].first, weeks[1].second).to_string()] = curr_days;
            }
            curr_days = {};
        }

        std::ofstream file("C://Scripts//final_json.json");
        file << std::setw(4) << json_data << std::endl;

        httplib::Client cli("http://10.99.26.80:8082");


        const std::string SECRET = "7777777";
        auto tokeExpiresAt = std::chrono::system_clock::now() + std::chrono::minutes(10);

        auto token = jwt::create()
            .set_type("JWS")
            .set_payload_claim("sched", jwt::claim(to_string(json_data)))
            .set_payload_claim("expires_at", jwt::claim(tokeExpiresAt))
            .sign(jwt::algorithm::hs256{ SECRET });

        httplib::Params params{
        { "jwtok", token }
        };

        auto res = cli.Post("/sched_update", params);
        std::cout << res->status << std::endl;
}

void timer_start(unsigned int interval) {
    std::thread([interval]() {
        while (true) {
            std::cout << "updating\n";
            auto x = std::chrono::steady_clock::now() + std::chrono::hours(interval);
            parsing();
            std::this_thread::sleep_until(x);
        }
        }).detach();
}


// Функция которая будет вызвана обработчиком, когда придёт запрос
void parsHandler(const Request& req, Response& res) {
    if (req.has_file("file")) {
        std::cout << "file has been found\n";
    }
    auto file = req.get_file_value("file");
    std::vector<unsigned char> file_content(file.content.begin(), file.content.end());
    std::istringstream stream(std::string(file_content.begin(), file_content.end()));
    xlnt::workbook wb;
    wb.load(stream);
    wb.save("C:\\Scripts\\file.xlsx");
    parsing();
    res.set_redirect("http://localhost:8080/parsOk", 200);
}

int main() {
    timer_start(168);
	Server svr;                  // Создаём сервер (пока-что не запущен)
	svr.Post("/pars", parsHandler);   
	svr.listen("0.0.0.0", 8083); // Запуск сервера на порту 8083
    
}