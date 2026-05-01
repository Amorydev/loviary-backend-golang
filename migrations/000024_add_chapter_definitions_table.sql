-- +goose Up
-- +goose StatementBegin

CREATE TABLE chapter_definitions (
  chapter_number   SMALLINT     PRIMARY KEY,
  title            VARCHAR(100) NOT NULL,
  subtitle         VARCHAR(200),
  day_start        INTEGER      NOT NULL,
  day_end          INTEGER,
  milestone_target INTEGER,
  cover_art_url    TEXT         DEFAULT 'https://thumbs.dreamstime.com/b/anime-landscape-lush-greenery-rocky-cliffs-distant-water-under-bright-sky-stunning-style-showcases-green-fields-vibrant-354691808.jpg',
  badge_icon_url   TEXT,
  theme_color      VARCHAR(7),
  created_at       TIMESTAMPTZ  DEFAULT NOW()
);

-- Mốc ngày: 30 | 100 | 180 (6th) | 365 (1yr) | 730 (2yr) | 1825 (5yr) | unlimited
INSERT INTO chapter_definitions
  (chapter_number, title, subtitle, day_start, day_end, milestone_target, theme_color)
VALUES
  (1, 'Chớm nở',            'Mọi thứ bắt đầu từ đây',               0,    30,   30,   '#FFB3C1'),
  (2, 'Thật sự cùng nhau',  'Khi hai người bắt đầu hiểu nhau hơn',  31,   100,  100,  '#FF8FAB'),
  (3, 'Nửa năm hạnh phúc',  'Sáu tháng — không phải ngẫu nhiên',    101,  180,  180,  '#E8637A'),
  (4, 'Tròn một năm yêu',   'Cột mốc đặc biệt đầu tiên',            181,  365,  365,  '#C9415A'),
  (5, 'Hai năm bên nhau',   'Từng ngày là một lựa chọn',             366,  730,  730,  '#A62042'),
  (6, 'Vững chắc theo năm', 'Năm năm — tình yêu đã được kiểm chứng',731,  1825, 1825, '#7A1530'),
  (7, 'Mãi mãi',            'Không có điểm kết',                     1826, NULL, NULL, '#4A0A1A');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS chapter_definitions CASCADE;

-- +goose StatementEnd
