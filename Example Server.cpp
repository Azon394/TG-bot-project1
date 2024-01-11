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

//ЧАСТЬ БОГДАНА НАЧАЛАСЬ

void parsing() {
    SetConsoleOutputCP(65001);

    json json_data;

    xlnt::workbook wb;
    wb.load("C:\\Scripts\\file.xlsx");
    ws = wb.sheet_by_index(1);

    auto rows_count = ws.calculate_dimension().height();
    auto cols_count = ws.calculate_dimension().width();

    std::vector <std::pair <size_t, size_t>> weeks;
    std::vector <std::pair <size_t, size_t>> groups;

    //Поиск групп и четной/нечетной недели
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
        size_t x = coordinates.first;  //группа по x
        size_t y = coordinates.second; //группа по y
        size_t week_num_x;
        size_t week_num_y;
        if (cnt <= (groups.size() / 2)) {
            week_num_x = weeks[0].first + 1;  // координаты x по неделе нечетной
            week_num_y = y + 1;               // координаты y по неделе нечетной
        }
        else {
            week_num_x = weeks[1].first + 1;  // координаты x по неделе четной
            week_num_y = y + 1;               // координаты y по неделе четной
        }
        size_t name_x = week_num_x - 1; //день недели
        size_t name_y = week_num_y;     //день недели
        size_t curr_num;        //текущий номер занятия
        int prev_num = -1;      //предыдущий номер занятия
        if (ws.cell(week_num_x, week_num_y).has_value()) {
            curr_num = stoi(ws.cell(week_num_x, week_num_y).to_string());
        }
        else {
            curr_num = prev_num;
        }
        std::vector<json> curr_lessons; //занятия
        std::queue<std::string> dates;  //дни недели
        while (week_num_y <= rows_count) {
            if (prev_num > curr_num && ws.cell(name_x, name_y).to_string() != "") {
                dates.push(ws.cell(name_x, name_y).to_string());
                if (!curr_lessons.empty()) {
                    curr_days.push_back({
                        {dates.front() , curr_lessons }
                        });
                    dates.pop();
                }
                curr_lessons = {};
            }
            if (prev_num != curr_num) {
                if (ws.cell(x, name_y).is_merged()) {
                    curr_lessons.push_back({
                    { "Название урока", ws.cell(x, name_y).to_string() },
                    { "Преподаватель", ws.cell(x, name_y + 1).to_string() },
                    { "Местонахождение", ws.cell(x, name_y + 2).to_string() },
                    { "Тип занятия", ws.cell(x - 1, name_y).to_string() },
                    { "Номер занятия", stoi(ws.cell(week_num_x, name_y).to_string()) },
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
                            { "Номер занятия", stoi(ws.cell(week_num_x, name_y).to_string()) },
                            { "Комментарий", ""},
                            { "Подгруппа",k + 1}
                                });
                        }
                        k++;
                    }
                }
            }
            week_num_y++;
            if (ws.cell(week_num_x, week_num_y).has_value()) {
                prev_num = curr_num;
                curr_num = stoi(ws.cell(week_num_x, week_num_y).to_string());
                name_y = week_num_y;
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

    httplib::Client cli("http://26.127.253.74:8082");


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
}


//ЧАСТЬ БОГДАНА ЗАКОНЧИЛАСЬ


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
    res.set_redirect("http://26.251.53.172:8080/parsOk", 200);
}

int main() {
    timer_start(168);
    Server svr;                  // Создаём сервер (пока-что не запущен)
    svr.Post("/pars", parsHandler);
    svr.listen("0.0.0.0", 8083); // Запуск сервера на порту 8083

}
