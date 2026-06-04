# Ruby app on a fat base, no cache mount.
FROM ruby:3.3
WORKDIR /app
COPY . .
RUN bundle install
CMD ["ruby", "app.rb"]
